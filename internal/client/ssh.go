package client

import (
	"bytes"
	"log"
	"net"
	"os"

	"github.com/inveracity/voki/internal/targets"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func TestConnection(config targets.Target, command string) (string, string, error) {
	sock, err := sshAgent()
	if err != nil {
		log.Fatalln(err)
	}

	cfg, err := makeSshConfig(config.User, sock)
	if err != nil {
		log.Fatalln(err)
	}

	sshclient, err := ssh.Dial("tcp", config.Host, cfg)
	if err != nil {
		log.Fatalln(err)
	}

	defer sshclient.Close()

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := sshclient.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	var e bytes.Buffer
	session.Stdout = &b
	session.Stderr = &e
	err = session.Run(command)

	return b.String(), e.String(), err
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
