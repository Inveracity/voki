package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/fatih/color"

	"github.com/inveracity/voki/internal/ssh"
	"github.com/inveracity/voki/internal/targets"
)

func (c *Client) Run(hcl string, username string) {
	config, err := targets.ParseHCL([]byte(hcl))
	if err != nil {
		log.Fatalln(err)
	}

	// Set the username
	if username != "" {
		for idx := range config.Targets {
			config.Targets[idx].User = &username
		}
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

			// Default to bash
			if step.Shell == "" {
				step.Shell = "bash"
			}
			cmd := fmt.Sprintf("%s -c '%s'", step.Shell, step.Command)

			// Run as sudo
			if step.Sudo {
				cmd = "sudo " + cmd
			}

			step.Command = cmd
			stdout, stderr, err := ssh.RunCommand(target, step.Command)

			fmt.Println("Result:")
			if err != nil {
				fmt.Fprintln(c.writer, color.RedString(stderr))
				log.Fatalln(err.Error())
			}
			fmt.Fprintln(c.writer, color.GreenString(stdout))

		// Copy a file to the remote server
		case "file":
			ctx := context.Background()
			fmt.Println("File", idx+1)
			fmt.Fprintln(c.writer, color.BlueString(step.Source))

			file := ssh.File{
				Source:      step.Source,
				Destination: step.Destination,
				Mode:        step.Mode,
			}

			temp, err := os.CreateTemp("", ".voki-*")
			defer temp.Close()
			if err != nil {
				log.Fatalln(err)
			}

			_, stderr, err := ssh.RunCommand(target, "mkdir -p "+temp.Name()+path.Dir(file.Destination))
			if err != nil {
				fmt.Fprintln(c.writer, color.RedString(stderr))
				log.Fatalln(err.Error())
			}

			ssh.TransferFile(ctx, *target.User, target.Host, file, temp.Name())

			_, stderr, err = ssh.RunCommand(target, "sudo mv "+temp.Name()+file.Destination+" "+file.Destination)
			if err != nil {
				fmt.Fprintln(c.writer, color.RedString(stderr))
				log.Fatalln(err.Error())
			}

			if step.Chown != "" {
				_, stderr, err := ssh.RunCommand(target, "sudo chown "+step.Chown+" "+step.Destination)
				if err != nil {
					fmt.Fprintln(c.writer, color.RedString(stderr))
					log.Fatalln(err.Error())
				}
			}

			_, stderr, err = ssh.RunCommand(target, "sudo rm -rf "+temp.Name())
			if err != nil {
				fmt.Fprintln(c.writer, color.RedString(stderr))
				log.Fatalln(err.Error())
			}
			fmt.Fprintln(c.writer, color.GreenString(step.Destination))

		// Parse a file with steps in it and run them
		case "task":
			// Recursively run a task
			config, err := targets.ParseHCL([]byte(step.Task))
			if err != nil {
				log.Fatalln(err)
			}
			c.ExecuteSteps(target, config.Steps)

		default:
			log.Fatalln("Unknown action", step.Action)
		}
	}
}
