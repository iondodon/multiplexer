package source

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iondodon/multiplexer/in/sshcon"
	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/out/client"
	"github.com/iondodon/multiplexer/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupConnection(t *testing.T) {
	mockClient := new(sshcon.MockSSHClient)
	mockSession := new(sshcon.MockSSHSession)
	mockQueue := new(queue.MockQueue)
	clientsCounter := client.ClientsCounter{
		NClients: &atomic.Uint64{},
		Cond:     sync.NewCond(&sync.Mutex{}),
	}

	mockClient.On("NewSession").Return(mockSession, nil)
	mockSession.On("StdoutPipe").Return(bytes.NewBuffer([]byte("Message1")), nil)
	mockSession.On("Start", producerScript).Return(nil)

	listener := &sourceListener{
		log:            logger.Get(),
		queue:          mockQueue,
		sshClient:      mockClient,
		clientsCounter: clientsCounter,
	}

	err := listener.SetupConnection()
	if err != nil {
		t.Errorf("no error was expected, got error: %v", err)
	}

	mockClient.AssertExpectations(t)
	mockSession.AssertExpectations(t)
}

func TestListen(t *testing.T) {
	mockClient := new(sshcon.MockSSHClient)
	mockSession := new(sshcon.MockSSHSession)
	mockQueue := new(queue.MockQueue)
	mockPipe := bytes.NewBuffer([]byte("Message1"))
	mockScanner := bufio.NewScanner(mockPipe)

	clientsCounter := client.ClientsCounter{
		NClients: &atomic.Uint64{},
		Cond:     sync.NewCond(&sync.Mutex{}),
	}
	clientsCounter.NClients.Store(1)

	listener := &sourceListener{
		log:            logger.Get(),
		queue:          mockQueue,
		sshClient:      mockClient,
		clientsCounter: clientsCounter,
		sshSession:     mockSession,
		scanner:        mockScanner,
	}

	mockSession.On("Close").Return(nil)
	mockQueue.On("Push", mock.Anything).Return()

	ctx, cancel := context.WithCancelCause(context.Background())
	listenErr, err := listener.Listen(ctx)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // give the listener 1 second to run

	cancel(errors.New("canceled by test"))

	err = <-listenErr
	expectedErrMsg := "finished scanning"
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("expected error message '%s', got '%v'", expectedErrMsg, err)
	}

	mockClient.AssertExpectations(t)
	mockSession.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
}
