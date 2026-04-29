package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/iondodon/multiplexer/internal/tcp"
)

func main() {
	ctx, cancelReceiving := context.WithCancel(context.Background())
	defer cancelReceiving()
	var dialer = net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:7070")
	if err != nil {
		slog.Error("failed to create connection", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		data, err := tcp.ReceiveNextFrame(conn)
		if err != nil {
			break
		}
		slog.Info("received", "frame", string(data))
	}
}
