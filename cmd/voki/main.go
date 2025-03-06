package main

import (
	"log"

	cmd "github.com/inveracity/voki/internal/cli"
)

func main() {
	cli := cmd.New()
	err := cli.Run()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}
