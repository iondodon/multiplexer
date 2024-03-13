package sshcon

import (
	"io"

	"github.com/stretchr/testify/mock"
)

type MockSSHClient struct {
	mock.Mock
}

func (msc *MockSSHClient) NewSession() (SSHSession, error) {
	args := msc.Called()
	return args.Get(0).(SSHSession), args.Error(1)
}

func (msc *MockSSHClient) Connect() error {
	args := msc.Called()
	return args.Error(0)
}

func (msc *MockSSHClient) Close() error {
	args := msc.Called()
	return args.Error(0)
}

type MockSSHSession struct {
	mock.Mock
}

func (msh *MockSSHSession) StdoutPipe() (io.Reader, error) {
	args := msh.Called()
	return args.Get(0).(io.Reader), args.Error(1)
}

func (msh *MockSSHSession) Start(cmd string) error {
	args := msh.Called(cmd)
	return args.Error(0)
}

func (msh *MockSSHSession) Close() error {
	args := msh.Called()
	return args.Error(0)
}
