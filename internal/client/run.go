package client

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"

	"github.com/inveracity/voki/internal/targets"
)

func (c *Client) Run(hcl string) {
	config, err := targets.ParseString([]byte(hcl))
	if err != nil {
		log.Fatalln(err)
	}

	for _, target := range config.Targets {
		fmt.Println("==== " + target.Name + " ====\n")

		c.ExecuteSteps(target, target.Steps)
	}
}

func (c *Client) ExecuteSteps(target targets.Target, steps []targets.Step) {
	for idx, step := range steps {
		switch step.Action {

		// Run commands on the remote server
		case "cmd":
			fmt.Println("Command", idx+1)
			fmt.Fprintln(c.writer, color.BlueString(step.Command))
			result := TestConnection(target, step.Command)
			fmt.Println("Result:")
			fmt.Fprintln(c.writer, color.GreenString(result))

		// Copy a file to the remote server
		case "file":
			ctx := context.Background()
			fmt.Println("File", idx+1)
			fmt.Fprintln(c.writer, color.BlueString(step.Source))

			file := File{
				Source:      step.Source,
				Destination: step.Destination,
				Mode:        step.Mode,
			}

			TransferFile(ctx, target.User, target.Host, file)

			fmt.Println("Result:")
			fmt.Fprintln(c.writer, color.GreenString(step.Destination))

		// Parse a file with steps in it and run them
		case "task":
			// Recursively run a task
			config, err := targets.ParseString([]byte(step.Task))
			if err != nil {
				log.Fatalln(err)
			}
			c.ExecuteSteps(target, config.Steps)

		default:
			log.Fatalln("Unknown action", step.Action)
		}
	}
}
