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
		m.serve()
	})
	m.wg.Wait()
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

func (m multiplexer) serve() {
	consumerListener, err := net.Listen("tcp", ":7070")
	if err != nil {
		slog.Error("failed to create connection listener", "error", err)
		os.Exit(1)
	}
	defer consumerListener.Close()

	for {
		conn, err := consumerListener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "error", err)
			continue
		}

		slog.Info("new connection", "connection", conn)
		go serveConsumer(conn, queue.GetReader())
	}
}

func serveConsumer(conn net.Conn, queueCursor *queue.Node) {
	defer conn.Close()
	for {
		data, next := queueCursor.ReadNext()
		if next != nil {
			if err := tcp.SendFrame(conn, []byte(data)); err != nil {
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
