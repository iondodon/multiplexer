package tcp

import (
	"encoding/binary"
	"io"
	"net"
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

func SendFrame(conn net.Conn, data []byte) error {
	var frameLengthBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(frameLengthBuf, uint32(len(data)))
	if err := writeFull(conn, frameLengthBuf); err != nil {
		return err
	}
	return writeFull(conn, data)
}
