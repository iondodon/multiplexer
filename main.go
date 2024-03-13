package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/iondodon/multiplexer/in/source"
	"github.com/iondodon/multiplexer/in/sshcon"
	"github.com/iondodon/multiplexer/logger"
	"github.com/iondodon/multiplexer/out/client"
	"github.com/iondodon/multiplexer/out/client/server"
	"github.com/iondodon/multiplexer/queue"
	"github.com/joho/godotenv"
)

const (
	failedLoafEnvFile        = "Error loading .env file: %v\n"
	failedConfigureSSHClient = "Failed to configure the SSH client: %v\n"
	failedStartTCPServer     = "Failed to start TCP server: %s\n"
	failedServerClients      = "Failed to serve clients: %s\n"
	failedStartListener      = "Failed to start listener: %s\n"
	failedSSHConnectionSetup = "Could not setup SSH connection: %s\n"
	failedConnectSSHClient   = "Failed to connect SSH client: %s\n"

	applicationStoppedCause = "Application stopped cause: %s\n"
	multiplexerHasStopped   = "Multiplexer has stopped"
)

func main() {
	logger := logger.Get()

	err := godotenv.Load()
	if err != nil {
		logger.Err.Fatalf(failedLoafEnvFile, err)
	}

	sshClient, err := sshcon.ConfigureNewClient()
	if err != nil {
		logger.Err.Fatalf(failedConfigureSSHClient, err)
	}
	err = sshClient.Connect()
	if err != nil {
		logger.Err.Fatalf(failedConnectSSHClient, err)
	}
	defer sshClient.Close()

	clientsCounter := client.ClientsCounter{
		NClients: &atomic.Uint64{},
		Cond:     sync.NewCond(&sync.Mutex{}),
	}
	queue := queue.New()

	sourceListener := source.NewListener(queue, clientsCounter, sshClient)
	err = sourceListener.SetupConnection()
	if err != nil {
		logger.Err.Fatalf(failedSSHConnectionSetup, err)
	}
	appContext, cancelApp := context.WithCancelCause(context.Background())
	listenErr, err := sourceListener.Listen(appContext)
	if err != nil {
		logger.Err.Fatalf(failedStartListener, err)
	}

	clientsServer := server.New(queue, clientsCounter)
	err = clientsServer.Connect()
	if err != nil {
		logger.Err.Fatalf(failedStartTCPServer, err)
	}
	defer clientsServer.Close()

	serveErr, err := clientsServer.Serve(appContext)
	if err != nil {
		logger.Err.Fatalf(failedServerClients, err)
	}

	close := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(close, os.Interrupt)

waitCloseEvent:
	for {
		select {
		case <-close:
			cancelApp(context.Canceled)
		case err := <-serveErr:
			cancelApp(err)
		case err := <-listenErr:
			cancelApp(err)
		case <-appContext.Done():
			break waitCloseEvent
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	appCancelErr := appContext.Err()
	if appCancelErr != nil && !errors.Is(appCancelErr, context.Canceled) {
		logger.Err.Printf(applicationStoppedCause, appCancelErr)
		os.Exit(1)
	}

	logger.Info.Println(multiplexerHasStopped)

	os.Exit(0)
}
