package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/iondodon/multiplexer/internal/tcp"
)

var counter uint64 = 0

func main() {
	ctx, cancelSending := context.WithCancel(context.Background())
	defer cancelSending()
	var dialer = net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "127.0.0.1:6060")
	if err != nil {
		slog.Error("failed to establish connection", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		frameData := []byte(strconv.FormatUint(counter, 10))
		err := tcp.SendFrame(conn, frameData)
		if err != nil {
			slog.Error("error sending frame", "error", err)
			break
		}
		slog.Info("sent", "data", counter)
		counter++
	}
}
