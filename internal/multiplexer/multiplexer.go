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
	wg                  *sync.WaitGroup
	proucerConnection   net.Conn
	consumerConnections []net.Conn
}

var m = multiplexer{
	wg:                  &sync.WaitGroup{},
	proucerConnection:   nil,
	consumerConnections: []net.Conn{},
}

func Get() multiplexer {
	return m
}

func (m multiplexer) Start() {
	m.wg.Go(func() {
		m.ingest()
	})
	m.wg.Go(func() {
		consumers := m.acceptConsumers()
		serveConsumers(consumers)
	})
	m.wg.Wait()
}

func serveConsumers(consumers <-chan net.Conn) {
	for consumer := range consumers {
		go serveConsumer(consumer)
	}
}

func serveConsumer(consumer net.Conn) {
	defer consumer.Close()
	queueCursor := queue.GetDedicatedReader()
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
}

func (m multiplexer) acceptConsumers() <-chan net.Conn {
	consumersListener, err := net.Listen("tcp", ":7070")
	if err != nil {
		slog.Error("failed to create connection listener", "error", err)
		os.Exit(1)
	}

	var consumers chan net.Conn = make(chan net.Conn)
	go func() {
		defer consumersListener.Close()
		for {
			conn, err := consumersListener.Accept()
			if err != nil {
				slog.Error("failed to accept connection", "error", err)
				continue
			}

			slog.Info("new connection", "connection", conn)
			consumers <- conn
		}
	}()

	return consumers
}

func (m multiplexer) ingest() {
	producerListener, err := net.Listen("tcp", ":6060")
	if err != nil {
		slog.Error("failed to create connection listener", "error", err)
		os.Exit(1)
	}
	defer producerListener.Close()

	conn, err := producerListener.Accept()
	if err != nil {
		slog.Error("failed to accept connection", "error", err)
		return
	}
	defer conn.Close()

	ingestStream(conn)
}

func ingestStream(conn net.Conn) {
	for {
		data, err := tcp.ReceiveNextFrame(conn)
		if err != nil {
			break
		}
		queue.Push(string(data))
	}
}
