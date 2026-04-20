package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
)

func receiveData(conn net.Conn) {
	defer conn.Close()
	var reader = bufio.NewReader(conn)
	var buffer = bytes.Buffer{}
	for {
		limitReader := io.LimitReader(reader, 1024)
		n, err := buffer.ReadFrom(limitReader)
		if n == 0 && err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("Connection closed by client")
			}
			slog.Error("Error reading data", "error", err)
			break
		}
		slog.Info("Data received", "data", buffer.String(), "bytesRead", n)
		buffer.Reset()
	}
}

func main() {
	listener, err := net.Listen("tcp", ":7070")
	if err != nil {
		slog.Error("Failer to create connection listener", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
		}
		go receiveData(connection)
	}
}
