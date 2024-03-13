package server

import (
	"context"
	"net"

	"github.com/stretchr/testify/mock"
)

type MockListener struct {
	mock.Mock
}

func (ml *MockListener) Accept() (net.Conn, error) {
	args := ml.Called()
	con, ok := args.Get(0).(net.Conn)
	if ok {
		return con, args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (ml *MockListener) Close() error {
	args := ml.Called()
	return args.Error(0)
}

func (ml *MockListener) Addr() net.Addr {
	args := ml.Called()
	return args.Error(0).(net.Addr)
}

type MockClientsServer struct {
	mock.Mock
}

func (mcs *MockClientsServer) Serve(ctx context.Context) error {
	args := mcs.Called()
	return args.Error(0)
}

func (mcs *MockClientsServer) Connect() error {
	args := mcs.Called()
	return args.Error(0)
}

func (mcs *MockClientsServer) Close() error {
	args := mcs.Called()
	return args.Error(0)
}
