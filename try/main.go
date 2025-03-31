package main

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/vault-client-go"
)

func main() {
	ctx := context.Background()

	// prepare a client with the given base address
	client, err := vault.New(
		vault.WithAddress("http://127.0.0.1:8200"),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	// authenticate with a root token (insecure)
	if err := client.SetToken("123456"); err != nil {
		log.Fatal(err)
	}

	// read the secret
	s, err := client.Secrets.KvV2Read(ctx, "voki", vault.WithMountPath("secret"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("secret retrieved:", s.Data.Data["test"])
}
