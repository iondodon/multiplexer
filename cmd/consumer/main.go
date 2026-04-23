package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
)

const maxFrameLength = 2048

func main() {
	conn, err := net.Dial("tcp", "localhost:7070")
	if err != nil {
		slog.Error("Failed to create connection", "error", err)
		os.Exit(1)
	}

	var frameLengthBuf = make([]byte, 4)
	for {
		n, err := io.ReadFull(conn, frameLengthBuf)
		if n == 0 && err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("Connection closed by client")
			} else if errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Warn("Connection closed while reading frame length")
			} else {
				slog.Error("Failed to read length prefix")
			}
			break
		}

		frameLength := binary.BigEndian.Uint32(frameLengthBuf)
		if frameLength > maxFrameLength {
			slog.Error("Frame length bigger than max allowed frame length")
			break
		}

		frameBuf := make([]byte, frameLength)
		n, err = io.ReadFull(conn, frameBuf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("Connection cloed by client")
			} else if errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Warn("Connection closed while reading frame length")
			} else {
				slog.Error("Failed to read frame", "error", err)
			}
			break
		}

		slog.Info("Received", "frame", string(frameBuf))
	}

}
