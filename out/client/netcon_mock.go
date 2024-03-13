package client

import (
	"net"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockConn struct {
	mock.Mock
}

func (mc *MockConn) Read(b []byte) (n int, err error) {
	args := mc.Called()
	return args.Int(0), args.Error(1)
}

func (mc *MockConn) Write(b []byte) (n int, err error) {
	args := mc.Called()
	return args.Int(0), args.Error(1)
}

func (mc *MockConn) Close() error {
	args := mc.Called()
	return args.Error(0)
}

func (mc *MockConn) LocalAddr() net.Addr {
	args := mc.Called()
	return args.Error(0).(net.Addr)
}

func (mc *MockConn) RemoteAddr() net.Addr {
	args := mc.Called()
	return args.Error(0).(net.Addr)
}

func (mc *MockConn) SetDeadline(t time.Time) error {
	args := mc.Called()
	return args.Error(0)
}

func (mc *MockConn) SetReadDeadline(t time.Time) error {
	args := mc.Called()
	return args.Error(0)
}

func (mc *MockConn) SetWriteDeadline(t time.Time) error {
	args := mc.Called()
	return args.Error(0)
}
