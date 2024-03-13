package feeder

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iondodon/multiplexer/in/source"
	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/out/client"
	"github.com/iondodon/multiplexer/queue"
	"github.com/stretchr/testify/mock"
)

func TestFeed(t *testing.T) {
	mockClient := new(client.MockConn)
	mockCursor := new(queue.MockCursor)

	nClients := new(atomic.Uint64)
	nClients.Store(1)

	mockClient.On("Write", mock.Anything).Return(source.ChunkSize, nil)
	mockClient.On("Close").Return(nil)
	mockCursor.On("Next").Return()
	msg := source.Fill0([]byte("1234567890"), source.ChunkSize)
	mockCursor.On("Get").Return(msg)
	mockCursor.On("HasNext").Return(true)

	clientFeeder := &clientFeeder{
		log:      logger.Get(),
		cursor:   mockCursor,
		nClients: nClients,
		client:   mockClient,
	}

	ctx, cancel := context.WithCancelCause(context.Background())

	go clientFeeder.Feed(ctx)

	time.Sleep(500 * time.Millisecond)

	cancel(errors.New("canceled by test"))

	time.Sleep(500 * time.Millisecond)

	mockClient.AssertExpectations(t)
	mockCursor.AssertExpectations(t)
}
