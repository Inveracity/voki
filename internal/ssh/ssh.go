package ssh

import (
	"bytes"
	"log"
	"net"
	"os"

	"github.com/inveracity/voki/internal/targets"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func RunCommand(sshclient *ssh.Client, command string) (string, string, error) {
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	session, err := sshclient.NewSession()
	defer session.Close()

	var b bytes.Buffer
	var e bytes.Buffer
	session.Stdout = &b
	session.Stderr = &e
	err = session.Run(command)

	return b.String(), e.String(), err
}

// CreateSSHClient creates a new SSH client using the provided target configuration.
func CreateSSHClient(config targets.Target) (*ssh.Client, error) {
	if config.Host == "localhost" {
		return nil, nil
	}

	sock, err := sshAgent()
	if err != nil {
		log.Fatalln(err)
	}

	cfg, err := makeSshConfig(*config.User, sock)
	if err != nil {
		log.Fatalln(err)
	}

	sshclient, err := ssh.Dial("tcp", config.Host, cfg)
	if err != nil {
		log.Fatalln(err)
	}

	return sshclient, nil
}

func sshAgent() (agent.ExtendedAgent, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}

	return agent.NewClient(sshAgent), nil
}

func makeSshConfig(user string, sock agent.ExtendedAgent) (*ssh.ClientConfig, error) {
	signers, err := sock.Signers()
	if err != nil {
		log.Printf("create signers error: %s", err)
		return nil, err
	}

	config := ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}

	return &config, nil
}
