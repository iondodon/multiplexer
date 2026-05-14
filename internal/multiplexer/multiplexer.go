package multiplexer

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"sync"
	"syscall"

	"github.com/iondodon/multiplexer/internal/queue"
	"github.com/iondodon/multiplexer/internal/tcp"
)

type multiplexer struct {
	wg *sync.WaitGroup
}

var m = multiplexer{
	wg: &sync.WaitGroup{},
}

func Get() multiplexer {
	return m
}

func (m multiplexer) Start() {
	consumers := m.acceptConsumers()
	producer := m.getProducer()

	m.serveConsumers(consumers)
	m.ingestStream(producer)

	m.wg.Wait()
}

func (m multiplexer) serveConsumers(consumers <-chan net.Conn) {
	m.wg.Go(func() {
		for consumer := range consumers {
			m.serveConsumer(consumer)
		}
	})
}

func (m multiplexer) serveConsumer(consumer net.Conn) {
	m.wg.Go(func() {
		queueCursor := queue.GetDedicatedReader()
		defer consumer.Close()
		for {
			data, next := queueCursor.Read()
			if next != nil {
				if err := tcp.SendFrame(consumer, []byte(data)); err != nil {
					if errors.Is(err, syscall.ECONNRESET) {
						slog.Info("peer reset connection")
						return
					}

					if errors.Is(err, syscall.EPIPE) {
						slog.Info("broken pipe / closed connection")
						return
					}

					var netErr net.Error
					if errors.As(err, &netErr) && netErr.Timeout() {
						slog.Info("network timeout")
						return
					}
				}
				queueCursor = next
			}
		}
	})
}

func (m multiplexer) acceptConsumers() <-chan net.Conn {
	consumersListener, err := net.Listen("tcp", ":7070")
	if err != nil {
		slog.Error("failed to create connection listener", "error", err)
		os.Exit(1)
	}

	var consumers chan net.Conn = make(chan net.Conn)
	m.wg.Go(func() {
		defer consumersListener.Close()
		// TODO: defer close consumers chan
		for {
			conn, err := consumersListener.Accept()
			if err != nil {
				slog.Error("failed to accept connection", "error", err)
				continue
			}

			slog.Info("new connection", "connection", conn)
			consumers <- conn
		}
	})

	return consumers
}

func (m multiplexer) getProducer() net.Conn {
	producerListener, err := net.Listen("tcp", ":6060")
	if err != nil {
		slog.Error("failed to create connection listener", "error", err)
		os.Exit(1)
	}
	defer producerListener.Close()

	conn, err := producerListener.Accept()
	if err != nil {
		slog.Error("failed to accept connection", "error", err)
		return nil
	}

	return conn
}

func (m multiplexer) ingestStream(conn net.Conn) {
	m.wg.Go(func() {
		for {
			data, err := tcp.ReceiveNextFrame(conn)
			if err != nil {
				break
			}
			queue.Push(data)
		}
	})
}
