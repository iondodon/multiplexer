package server

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/out/client"
	"github.com/iondodon/multiplexer/queue"
)

func TestServer_AcceptsConnectionAndGetsNextMessage(t *testing.T) {
	mockListener := new(MockListener)
	mockConn := new(client.MockConn)
	mockQueue := new(queue.MockQueue)
	mockCursor := new(queue.MockCursor)
	clientsCounter := client.ClientsCounter{
		NClients: &atomic.Uint64{},
		Cond:     sync.NewCond(&sync.Mutex{}),
	}

	mockQueue.On("GetCursor").Return(mockCursor)
	mockCursor.On("HasNext").Return(false)
	mockListener.On("Accept").Return(mockConn, nil)
	mockConn.On("Close").Return(nil)

	clientsServer := &clientsServer{
		log:            logger.Get(),
		queue:          mockQueue,
		clientsCounter: clientsCounter,
		listener:       mockListener,
	}

	appContext, cancel := context.WithCancelCause(context.Background())
	_, err := clientsServer.Serve(appContext)
	if err != nil {
		t.Errorf("no error message expected, got %s", err)
	}

	cancel(errors.New("canceled by test"))
}

func TestServer(t *testing.T) {
	mockListener := new(MockListener)
	mockQueue := new(queue.MockQueue)
	clientsCounter := client.ClientsCounter{
		NClients: &atomic.Uint64{},
		Cond:     sync.NewCond(&sync.Mutex{}),
	}

	mockListener.On("Close").Return(nil)

	clientsServer := &clientsServer{
		log:            logger.Get(),
		queue:          mockQueue,
		clientsCounter: clientsCounter,
	}

	appContext, _ := context.WithCancelCause(context.Background())
	_, err := clientsServer.Serve(appContext)
	if err == nil {
		t.Errorf("expected error, no error returned instead")
	}
}

func TestServer_ServeErrorOccured(t *testing.T) {
	mockListener := new(MockListener)
	mockQueue := new(queue.MockQueue)
	clientsCounter := client.ClientsCounter{
		NClients: &atomic.Uint64{},
		Cond:     sync.NewCond(&sync.Mutex{}),
	}

	mockListener.On("Accept").Return(nil, errors.New("error occurred while accepting"))
	mockListener.On("Close").Return(nil)

	clientsServer := &clientsServer{
		log:            logger.Get(),
		queue:          mockQueue,
		clientsCounter: clientsCounter,
		listener:       mockListener,
	}

	appContext, _ := context.WithCancelCause(context.Background())
	serverErr, err := clientsServer.Serve(appContext)
	if err != nil {
		t.Errorf("no error expected, got %v: ", err)
	}

	err = <-serverErr
	if err == nil {
		t.Errorf("expected serve error, no error returned instead")
	}
}
