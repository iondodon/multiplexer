package main

import (
	"log/slog"
	"net"
	"os"
	"sync"

	tcp "github.com/iondodon/multiplexer/internal"
	"github.com/iondodon/multiplexer/internal/queue"
)

func receiveData(conn net.Conn) {
	defer conn.Close()
	for {
		data, err := tcp.Receive(conn)
		if err != nil {
			break
		}
		queue.Push(string(data))
	}
}

func serveConsumer(conn net.Conn, reader *queue.Node) {
	for {
		data, next := reader.ReadNext()
		if next != nil {
			tcp.SendFrame(conn, []byte(data))
			reader = next
		}
	}
}

func main() {
	var wg = &sync.WaitGroup{}

	wg.Go(func() {
		producerListener, err := net.Listen("tcp", ":6060")
		if err != nil {
			slog.Error("Failed to create connection listener", "error", err)
			os.Exit(1)
		}
		defer producerListener.Close()

		for {
			conn, err := producerListener.Accept()
			if err != nil {
				slog.Error("Failed to accept connection", "error", err)
			} else {
				go receiveData(conn)
			}
		}
	})

	wg.Go(func() {
		consumerListener, err := net.Listen("tcp", ":7070")
		if err != nil {
			slog.Error("Failed to create connection listener", "error", err)
			os.Exit(1)
		}
		defer consumerListener.Close()

		for {
			conn, err := consumerListener.Accept()
			if err != nil {
				slog.Error("Failed to accept connection", "error", err)
			} else {
				slog.Info("New connection", "connection", conn)
			}

			var reader = queue.GetReader()
			go serveConsumer(conn, reader)
		}
	})

	wg.Wait()

}
