package client

import (
	"fmt"

	"github.com/inveracity/voki/internal/targets"
)

type Client struct {
}

func New() *Client {
	return &Client{}
}

func (c *Client) Ping(targetname string) {
	targetconfig := targets.Parse()

	for _, target := range targetconfig {
		if targetname != "" {
			if target.Name == targetname {
				TestConnection(target)
				fmt.Println(target.Name, target.Host, "is reachable")
			}
		} else {
			TestConnection(target)
			fmt.Println(target.Name, target.Host, "is reachable")
		}
	}
}
