package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/iondodon/multiplexer/internal/queue"
)

type Pusher interface {
	Push(data string)
}

const maxFrameLength = 2048

func receiveData(conn net.Conn, pusher Pusher) {
	defer conn.Close()

	var lengthBuf = make([]byte, 8)
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

		frameLength := binary.BigEndian.Uint64(lengthBuf)
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

		slog.Info("Data received", "length", frameLength, "data", string(frameBuf))
		pusher.Push(string(frameBuf))
	}
}

func main() {
	var q Pusher = queue.GetInstance()

	listener, err := net.Listen("tcp", ":7070")
	if err != nil {
		slog.Error("Failed to create connection listener", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
		}
		go receiveData(connection, q)
	}
}
