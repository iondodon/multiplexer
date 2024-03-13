package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/out/client"
	"github.com/iondodon/multiplexer/out/client/feeder"
	"github.com/iondodon/multiplexer/queue"
)

const (
	multiplexerServing                    = "Multiplexer serving on port %v\n"
	serverNotConnected                    = "server is not connected; call Connect() to start a connection"
	failedAcceptConnection                = "Failed to accept connection: %s\n"
	temporaryError                        = "Temporary error occurred while attempting to accept a new connection: %s\n"
	exitingCauseFailedAcceptNewConnection = "exiting cause failed to accept new connection: %w"
)

const (
	tcpListenPort = ":8080"
	protocol      = "tcp"
)

type ServeError chan error

type ClientsServer interface {
	Serve(ctx context.Context) (ServeError, error)
	Connect() error
	Close() error
}

type clientsServer struct {
	log            logger.Logger
	queue          queue.Queue
	clientsCounter client.ClientsCounter
	listener       net.Listener
}

func New(queue queue.Queue, clientsCounter client.ClientsCounter) ClientsServer {
	return &clientsServer{
		log:            logger.Get(),
		queue:          queue,
		clientsCounter: clientsCounter,
	}
}

func (cs *clientsServer) Connect() error {
	listener, err := net.Listen(protocol, tcpListenPort)
	if err != nil {
		return err
	}
	cs.listener = listener

	cs.log.Info.Printf(multiplexerServing, tcpListenPort)

	return nil
}

func (cs *clientsServer) Serve(appContext context.Context) (ServeError, error) {
	if cs.listener == nil {
		return nil, errors.New(serverNotConnected)
	}

	var serveErr ServeError = make(ServeError, 1)
	go func() {
		for {
			client, err := cs.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Temporary() {
					cs.log.Err.Printf(temporaryError, err)
					continue
				}
				cs.log.Err.Printf(failedAcceptConnection, err)
				cs.listener.Close()
				serveErr <- fmt.Errorf(exitingCauseFailedAcceptNewConnection, err)
				close(serveErr)
				break
			}

			cs.clientsCounter.NClients.Add(1)

			if cs.clientsCounter.NClients.Load() == 1 {
				cs.clientsCounter.Cond.L.Lock()
				cs.clientsCounter.Cond.Signal()
				cs.clientsCounter.Cond.L.Unlock()
			}

			clientFeeder := feeder.New(cs.queue.GetCursor(), cs.clientsCounter.NClients, client)
			go clientFeeder.Feed(appContext)
		}
	}()

	return serveErr, nil
}

func (cs *clientsServer) Close() error {
	if cs.listener != nil {
		return cs.listener.Close()
	}
	return nil
}
