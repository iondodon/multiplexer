package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"

	tcp "github.com/iondodon/multiplexer/internal"
	"github.com/iondodon/multiplexer/internal/queue"
)

const maxFrameLength = 2048

func receiveData(conn net.Conn) {
	defer conn.Close()

	var lengthBuf = make([]byte, 4)
	for {
		n, err := io.ReadFull(conn, lengthBuf)
		if n == 0 && err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("Connection closed by client")
			} else if errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Warn("Connection closed while reading frame length")
			} else {
				slog.Error("Error reading frame length", "error", err)
			}
			break
		}

		frameLength := binary.BigEndian.Uint32(lengthBuf)
		if frameLength > maxFrameLength {
			slog.Error("Frame length bigger than max allowed frame length")
			break
		}

		frameBuf := make([]byte, frameLength)
		n, err = io.ReadFull(conn, frameBuf)
		if n == 0 && err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Warn("Connection closed while reading frame", "expected", frameLength, "bytesRead", n)
			} else {
				slog.Error("Error reading frame", "error", err, "bytesRead", n)
			}
			break
		}

		queue.Push(string(frameBuf))
	}
}

func sendToConsumer(conn net.Conn, reader *queue.Node) {
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
			go sendToConsumer(conn, reader)
		}
	})

	wg.Wait()

}
