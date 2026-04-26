package tcp

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"syscall"
)

const maxFrameLength = 2048

func ReadNextFrame(conn net.Conn) ([]byte, error) {
	var lengthBuf = make([]byte, 4)
	n, err := io.ReadFull(conn, lengthBuf)
	if n == 0 && err != nil {
		if errors.Is(err, io.EOF) {
			slog.Info("Connection closed by client")
		} else if errors.Is(err, io.ErrUnexpectedEOF) {
			slog.Warn("Connection closed while reading frame length")
		} else {
			slog.Error("Error reading frame length", "error", err)
		}
		return nil, err
	}

	frameLength := binary.BigEndian.Uint32(lengthBuf)
	if frameLength > maxFrameLength {
		slog.Error("Frame length bigger than max allowed frame length")
		return nil, errors.New("Frame length bigger than max allowed frame lengt")
	}

	frameBuf := make([]byte, frameLength)
	n, err = io.ReadFull(conn, frameBuf)
	if n == 0 && err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			slog.Warn("Connection closed while reading frame", "expected", frameLength, "bytesRead", n)
		} else {
			slog.Error("Error reading frame", "error", err, "bytesRead", n)
		}
		return nil, err
	}

	return frameBuf, nil
}

func SendFrame(conn net.Conn, data []byte) error {
	var frameLengthBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(frameLengthBuf, uint32(len(data)))
	if err := writeFull(conn, frameLengthBuf); err != nil {
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

		return err
	}
	return writeFull(conn, data)
}

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
