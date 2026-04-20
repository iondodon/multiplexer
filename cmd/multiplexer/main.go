package main

import (
	"bufio"
	"bytes"
	"log/slog"
	"net"
	"os"
)

func receiveData(conn net.Conn) {
	defer conn.Close()
	var reader = bufio.NewReader(conn)
	var buffer = bytes.NewBuffer([]byte{})
	for {
		n, err := buffer.ReadFrom(reader)
		if err != nil {
			slog.Error("Error reading data", "error", err)
			break
		}
		slog.Info("Data received", "data", buffer.Bytes(), "bytesRead", n)
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
