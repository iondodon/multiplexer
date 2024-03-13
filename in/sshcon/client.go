package sshcon

import (
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/iondodon/multiplexer/logger"
)

const protocol = "tcp"

const (
	missingEnvVar       = "%s environment variable is not present"
	failedSSHConnection = "failed to start a client connection to the given SSH server"
	failedCreateSession = "Failed to create session: %s\n"
	clientNotConfigured = "client not configured"
	missingSSHClient    = "missing SSH client"
)

type SSHSession interface {
	StdoutPipe() (io.Reader, error)
	Start(cmd string) error
	Close() error
}

type SSHClient interface {
	NewSession() (SSHSession, error)
	Connect() error
	Close() error
}

type Config struct {
	clientConfig *ssh.ClientConfig
	sshAddr      string
}

type sshClient struct {
	log    logger.Logger
	config *Config
	client *ssh.Client
}

func ConfigureNewClient() (SSHClient, error) {
	username, isPresent := os.LookupEnv("USERNAME")
	if !isPresent {
		return nil, fmt.Errorf(missingEnvVar, "USERNAME")
	}
	password, isPresent := os.LookupEnv("PASSWORD")
	if !isPresent {
		return nil, fmt.Errorf(missingEnvVar, "PASSWORD")
	}
	sshAddr, isPResent := os.LookupEnv("SSH_ADDRESS")
	if !isPResent {
		return nil, fmt.Errorf(missingEnvVar, "SSH_ADDRESS")
	}

	clientConfig := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Should not be used for production
	}

	return &sshClient{
		log: logger.Get(),
		config: &Config{
			clientConfig: clientConfig,
			sshAddr:      sshAddr,
		},
	}, nil
}

func (c *sshClient) Connect() error {
	if c.config == nil {
		return errors.New(clientNotConfigured)
	}

	client, err := ssh.Dial(protocol, c.config.sshAddr, c.config.clientConfig)
	if err != nil {
		return errors.New(failedSSHConnection)
	}
	c.client = client

	return nil
}

func (c *sshClient) NewSession() (SSHSession, error) {
	if c.client == nil {
		return nil, errors.New(missingSSHClient)
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (c *sshClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
