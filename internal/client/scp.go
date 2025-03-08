package client

import (
	"context"
	"log"
	"os"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"golang.org/x/crypto/ssh"
)

type File struct {
	Source      string
	Destination string
	Mode        string
}

// Transfer a file to a remote server
func TransferFile(ctx context.Context, user, host string, file File) {
	clientConfig, _ := auth.SshAgent(user, ssh.InsecureIgnoreHostKey())

	client := scp.NewClient(host, &clientConfig)

	err := client.Connect()
	if err != nil {
		log.Fatalln("Couldn't establish a connection to the remote server ", err)
	}

	f, _ := os.Open(file.Source)
	defer client.Close()
	defer f.Close()

	err = client.CopyFromFile(ctx, *f, file.Destination, file.Mode)

	if err != nil {
		log.Fatalln("Error while copying file ", err)
	}
}
