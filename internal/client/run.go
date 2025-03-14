package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/inveracity/voki/internal/print"
	"github.com/inveracity/voki/internal/ssh"
	"github.com/inveracity/voki/internal/targets"
)

func (c *Client) Run(hcl string, username string) {
	config, err := targets.ParseHCL([]byte(hcl), c.EvalContext)
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
		print.Title(target.Name)

		c.ExecuteSteps(target, target.Steps)
	}
}

func (c *Client) ExecuteSteps(target targets.Target, steps []targets.Step) {
	for _, step := range steps {
		switch step.Action {

		// Run commands on the remote server
		case "cmd":
			fmt.Println("Command:")
			print.Info(step.Command)

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
				print.Error(stderr)
				print.Fatal(err)
			}

			print.Success(stdout)

		// Copy a file to the remote server
		case "file":
			ctx := context.Background()
			fmt.Println("File:")
			temp, err := os.CreateTemp("", ".voki-*")
			defer temp.Close()
			if err != nil {
				log.Fatalln(err)
			}

			file := ssh.File{
				Source:      step.Source,
				Destination: step.Destination,
				Mode:        step.Mode,
			}

			if step.Source != "" {
				print.Info(step.Source)
			}

			if step.Data != "" && step.Source == "" {
				print.Info(step.Data)
				if err := os.WriteFile(temp.Name()+"render.tmp", []byte(step.Data), 0644); err != nil {
					log.Fatalln(err)
				}
				file.Source = temp.Name() + "render.tmp"
			}

			_, stderr, err := ssh.RunCommand(target, "mkdir -p "+temp.Name()+path.Dir(file.Destination))
			if err != nil {
				print.Error(stderr)
				print.Fatal(err)
			}

			ssh.TransferFile(ctx, *target.User, target.Host, file, temp.Name())

			fmt.Println("Result:")

			_, stderr, err = ssh.RunCommand(target, "sudo mv "+temp.Name()+file.Destination+" "+file.Destination)
			if err != nil {
				print.Error(stderr)
				print.Fatal(err)
			}

			if step.Chown != "" {
				_, stderr, err := ssh.RunCommand(target, "sudo chown "+step.Chown+" "+step.Destination)
				if err != nil {
					print.Error(stderr)
					print.Fatal(err)
				}
			}

			_, stderr, err = ssh.RunCommand(target, "sudo rm -rf "+temp.Name())
			if err != nil {
				print.Error(stderr)
				print.Fatal(err)
			}

			print.Success(step.Destination)

		// Parse a file with steps in it and run them
		case "task":
			// Recursively run a task
			config, err := targets.ParseHCL([]byte(step.Task), c.EvalContext)
			if err != nil {
				print.Fatal(err)
			}
			c.ExecuteSteps(target, config.Steps)

		default:
			log.Fatalln("Unknown action", step.Action)
		}
	}
}
