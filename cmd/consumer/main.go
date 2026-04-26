package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	tcp "github.com/iondodon/multiplexer/internal"
)

func main() {
	ctx, cancelReceiving := context.WithCancel(context.Background())
	defer cancelReceiving()
	var dialer = net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:7070")
	if err != nil {
		slog.Error("Failed to create connection", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		data, err := tcp.ReadNextFrame(conn)
		if err != nil {
			break
		}
		slog.Info("Received", "frame", string(data))
	}
}
