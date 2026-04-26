package main

import (
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/iondodon/multiplexer/internal/queue"
	"github.com/iondodon/multiplexer/internal/tcp"
)

func main() {
	var wg = &sync.WaitGroup{}

	wg.Go(func() {
		producerListener, err := net.Listen("tcp", ":6060")
		if err != nil {
			slog.Error("failed to create connection listener", "error", err)
			os.Exit(1)
		}
		defer producerListener.Close()

		for {
			conn, err := producerListener.Accept()
			if err != nil {
				slog.Error("failed to accept connection", "error", err)
			} else {
				go ingestStream(conn)
			}
		}
	})

	wg.Go(func() {
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
			} else {
				slog.Info("new connection", "connection", conn)
			}

			var reader = queue.GetReader()
			go serveConsumer(conn, reader)
		}
	})

	wg.Wait()

}

func ingestStream(conn net.Conn) {
	defer conn.Close()
	for {
		data, err := tcp.ReadNextFrame(conn)
		if err != nil {
			break
		}
		queue.Push(string(data))
	}
}

func serveConsumer(conn net.Conn, queueCursor *queue.Node) {
	for {
		data, next := queueCursor.ReadNext()
		if next != nil {
			tcp.SendFrame(conn, []byte(data))
			queueCursor = next
		}
	}
}
