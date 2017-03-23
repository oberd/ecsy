package ssh

import (
	"fmt"

	"github.com/glinton/ssh"
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

// func privateKeyFile(file string) ssh.AuthMethod {
// 	buffer, err := ioutil.ReadFile(file)
// 	if err != nil {
// 		return nil
// 	}
//
// 	key, err := ssh.ParsePrivateKey(buffer)
// 	if err != nil {
// 		return nil
// 	}
// 	return ssh.PublicKeys(key)
// }

// Connect will try and connect up an SSH session
func (client *Client) Connect() error {
	keys := ssh.Auth{
		Keys: []string{client.PrivateKeyFile},
	}
	sshterm, err := ssh.NewNativeClient(client.User, client.Host, "SSH-2.0-MyCustomClient-1.0", client.Port, &keys)
	err = sshterm.Shell()
	if err != nil && err.Error() != "exit status 255" {
		return fmt.Errorf("Failed to request shell - %s", err)
	}
	return nil
	// sshConfig := &ssh.ClientConfig{
	// 	User: client.User,
	// 	Auth: []ssh.AuthMethod{
	// 		privateKeyFile(client.PrivateKeyFile),
	// 	},
	// }
	// connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port), sshConfig)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to dial: %s", err)
	// }
	// session, err := connection.NewSession()
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to create session: %s", err)
	// }
	// defer session.Close()
	// session.Stdout = os.Stdout
	// session.Stderr = os.Stderr
	// in, _ := session.StdinPipe()
	//
	// // Set up terminal modes
	// modes := ssh.TerminalModes{
	// 	ssh.ECHO:          0,     // disable echoing
	// 	ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
	// 	ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	// }
	//
	// // Request pseudo terminal
	// if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
	// 	return nil, fmt.Errorf("request for pseudo terminal failed: %s", err)
	// }
	//
	// // Start remote shell
	// if err := session.Shell(); err != nil {
	// 	return nil, fmt.Errorf("failed to start shell: %s", err)
	// }
	//
	// // Accepting commands
	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	str, _ := reader.ReadString('\n')
	// 	fmt.Fprint(in, str)
	// 	if str == "exit\n" {
	// 		return nil, nil
	// 	}
	// }
}
