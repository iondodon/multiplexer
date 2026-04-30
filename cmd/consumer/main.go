package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/iondodon/multiplexer/internal/tcp"
)

var shouldBe uint64 = 0

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

		numberStr := string(data)
		slog.Info("received", "frame", string(data))

		number, err := strconv.ParseUint(numberStr, 10, 64)
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		if shouldBe > 0 && number != shouldBe {
			slog.Error("HOPA", "number", number, "shouldBe", shouldBe)
			break
		}
		shouldBe = number + 1
	}
}
