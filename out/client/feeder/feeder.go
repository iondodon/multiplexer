package feeder

import (
	"context"
	"errors"
	"net"
	"sync/atomic"

	"github.com/iondodon/multiplexer/in/source"
	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/queue"
)

const (
	canceledClientFeeder = "Client feeder has been canceled"
	stopedClientFeeder   = "Client feeder stoped because an error occurred: %s\n"
	couldNotSendChunk    = "could not send full chunk successfully"
)

const writeFailsLimit = 3

type ClientFeeder interface {
	Feed(appContext context.Context)
}

type clientFeeder struct {
	log      logger.Logger
	cursor   queue.Cursor
	nClients *atomic.Uint64
	client   net.Conn
}

func New(cursor queue.Cursor, nClients *atomic.Uint64, conn net.Conn) ClientFeeder {
	return &clientFeeder{
		log:      logger.Get(),
		cursor:   cursor,
		nClients: nClients,
		client:   conn,
	}
}

func (cf *clientFeeder) Feed(appContext context.Context) {
	defer cf.client.Close()

feed:
	for {
		select {
		case <-appContext.Done():
			cf.log.Info.Println(canceledClientFeeder)
			break feed
		default:
			if !cf.cursor.HasNext() {
				continue
			}
			err := cf.sendNextMessage()
			if err != nil {
				cf.nClients.Add(^uint64(0))
				break feed
			}
		}
	}
}

func (cf *clientFeeder) sendNextMessage() error {
	cf.cursor.Next()

	message := cf.cursor.Get()
	fails := 0
	for i := 0; fails < writeFailsLimit && i < len(message); {
		chunk := message[i : i+source.ChunkSize]
		bytesWriten, err := cf.client.Write(chunk)
		if err != nil {
			cf.log.Err.Printf(stopedClientFeeder, err)
			return err
		}
		if bytesWriten != source.ChunkSize {
			// the receiving side should discard received chunks
			// that have size != chunkSize
			// and should try to listen again for the same chunk
			fails++
			continue
		}
		i += source.ChunkSize
	}

	if fails == writeFailsLimit {
		return errors.New(couldNotSendChunk)
	}

	return nil
}
