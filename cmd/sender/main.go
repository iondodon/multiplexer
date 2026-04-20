package main

import (
	"context"
	"log/slog"
	"net"
	"os"
)

func main() {
	ctx, cancelSendig := context.WithCancel(context.Background())
	defer cancelSendig()
	var dialer = net.Dialer{}
	connection, err := dialer.DialContext(ctx, "tcp", ":7070")
	if err != nil {
		slog.Error("Failer to establish connection", "error", err)
		os.Exit(1)
	}
	defer connection.Close()

	for {
		_, err := connection.Write([]byte("Hi"))
		if err != nil {
			slog.Error("Error writing bytes", "error", err.Error())
		}
	}
}
