package main

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
)

// For write there is no io.WriteFull (equivalent to the existing io.ReadFull)
// that would guarantee that the entire message was writen.
// The user is responsable for handling this.
func writeFull(conn net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrUnexpectedEOF
		}
		data = data[n:]
	}
	return nil
}

func sendFrame(conn net.Conn, data []byte) error {
	var lengthPrefixBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(lengthPrefixBuf, uint32(len(data)))
	err := writeFull(conn, lengthPrefixBuf)
	if err != nil {
		return err
	}

	return writeFull(conn, data)
}

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

	var counter uint64 = 0
	for {
		frameData := []byte(strconv.FormatUint(counter, 10))
		err := sendFrame(conn, frameData)
		if err != nil {
			slog.Error("Error sending frame", "error", err)
			break
		}
		counter++
	}
}
