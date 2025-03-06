package client

import (
	"fmt"

	"github.com/inveracity/voki/internal/targets"
)

func (c *Client) Run(specfile string) {
	config := targets.Parse(specfile)

	for _, target := range config.Targets {
		fmt.Println("target: " + target.Name)
		TestConnection(target)
	}

}
