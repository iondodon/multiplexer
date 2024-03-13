package source

import (
	"bufio"
	"context"
	"errors"
	"fmt"

	"github.com/iondodon/multiplexer/in/sshcon"
	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/out/client"
	"github.com/iondodon/multiplexer/queue"
)

const producerScript = `
	counter=0
	while true; do
		let counter+=1
		echo "Message $counter"
	done
`

const (
	failedGetSSHSession = "failed to get SSH session: %w"
	failedGetStdoutPipe = "failed to get stdout pipe: %w"
	failedStartCommand  = "failed to start command: %w"
	failedReadFromPipe  = "error reading from pipe: %w"
	stoppedMessageSend  = "finished scanning"
	connectionNotSet    = "connection is not setup"

	canceledScanning = "Messages scanning has been canceled"
)

type ListenError chan error

type SourceListener interface {
	Listen(appContext context.Context) (ListenError, error)
	SetupConnection() error
}

type sourceListener struct {
	log logger.Logger

	queue          queue.Queue
	clientsCounter client.ClientsCounter

	sshClient  sshcon.SSHClient
	scanner    *bufio.Scanner
	sshSession sshcon.SSHSession
}

func NewListener(queue queue.Queue, clientsCounter client.ClientsCounter, sshClient sshcon.SSHClient) SourceListener {
	return &sourceListener{
		log:            logger.Get(),
		queue:          queue,
		clientsCounter: clientsCounter,
		sshClient:      sshClient,
	}
}

func (sl *sourceListener) SetupConnection() error {
	sshSession, err := sl.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf(failedGetSSHSession, err)
	}
	sl.sshSession = sshSession

	pipe, err := sshSession.StdoutPipe()
	if err != nil {
		return fmt.Errorf(failedGetStdoutPipe, err)
	}

	if err := sshSession.Start(producerScript); err != nil {
		return fmt.Errorf(failedStartCommand, err)
	}
	scanner := bufio.NewScanner(pipe)
	sl.scanner = scanner

	return nil
}

func (sl *sourceListener) Listen(appContext context.Context) (ListenError, error) {
	if sl.sshSession == nil || sl.scanner == nil {
		return nil, errors.New(connectionNotSet)
	}

	var listenErr ListenError = make(ListenError, 1)
	go func() {
		defer sl.sshSession.Close()
	listen:
		for {
			select {
			case <-appContext.Done():
				sl.log.Info.Println(canceledScanning)
				break listen
			default:
				sl.clientsCounter.Cond.L.Lock()
				if sl.clientsCounter.NClients.Load() == 0 {
					sl.clientsCounter.Cond.Wait()
				}
				sl.clientsCounter.Cond.L.Unlock()
				if sl.scanner.Scan() {
					msg := Fill0(sl.scanner.Bytes(), ChunkSize)
					sl.queue.Push(msg)
				} else {
					break listen
				}
			}
		}

		if err := sl.scanner.Err(); err != nil {
			listenErr <- fmt.Errorf(failedReadFromPipe, err)
			close(listenErr)
			return
		}

		listenErr <- errors.New(stoppedMessageSend)
	}()

	return listenErr, nil
}
