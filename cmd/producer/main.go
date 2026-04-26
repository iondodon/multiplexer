package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"strconv"
	"syscall"

	tcp "github.com/iondodon/multiplexer/internal"
)

var counter uint64 = 0

func main() {
	ctx, cancelSending := context.WithCancel(context.Background())
	defer cancelSending()
	var dialer = net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "127.0.0.1:6060")
	if err != nil {
		slog.Error("Failed to establish connection", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	for {
		frameData := []byte(strconv.FormatUint(counter, 10))
		err := tcp.SendFrame(conn, frameData)
		if err != nil {
			if errors.Is(err, syscall.ECONNRESET) {
				slog.Info("peer reset connection")
			}

			if errors.Is(err, syscall.EPIPE) {
				slog.Info("broken pipe / closed connection")
			}

			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				slog.Info("network timeout")
			}

			slog.Error("Error sending frame", "error", err)

			break
		}
		slog.Info("Sent", "data", counter)
		counter++
	}
}
