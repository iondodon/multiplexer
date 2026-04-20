package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strconv"
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

	var counter uint64 = 0
	for {
		_, err := connection.Write([]byte(strconv.FormatUint(counter, 10)))
		if err != nil {
			slog.Error("Error writing bytes", "error", err.Error())
		}
		counter++
	}
}
