package ssh

import (
	"fmt"

	"github.com/nnnnathann/ssh"
)

// Client is used to connect over ssh to a remote host.
type Client struct {
	Host           string
	Port           int
	User           string
	PrivateKeyFile string
}

// ClientConfiguration is used to create a client
type ClientConfiguration struct {
	Host           string
	Port           int
	User           string
	PrivateKeyFile string
}

// NewClient is used to create a new client!
func NewClient(config ClientConfiguration) *Client {
	out := &Client{
		Port: 22,
		User: "ec2-user",
	}
	if config.Host != "" {
		out.Host = config.Host
	}
	if config.Port != 0 {
		out.Port = config.Port
	}
	if config.User != "" {
		out.User = config.User
	}
	if config.PrivateKeyFile != "" {
		out.PrivateKeyFile = config.PrivateKeyFile
	}
	return out
}

// Connect will try and connect up an SSH session
func (client *Client) Connect() error {
	keys := ssh.Auth{
		Keys: []string{client.PrivateKeyFile},
	}
	sshterm, err := ssh.NewNativeClient(client.User, client.Host, "SSH-2.0-MyCustomClient-1.0", client.Port, &keys, nil)
	if err != nil {
		return fmt.Errorf("Failed to request shell - %s", err)
	}
	err = sshterm.Shell()
	if err != nil && err.Error() != "exit status 255" {
		return fmt.Errorf("Failed to request shell - %s", err)
	}
	return nil
}
