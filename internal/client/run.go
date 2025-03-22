package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/inveracity/voki/internal/ssh"
	"github.com/inveracity/voki/internal/targets"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
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
		c.Printer.Title(target.Name)

		bar := c.Bar.AddBar(0,
			mpb.BarOptional(mpb.BarWidth(40), true),
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("%-27s", target.Name)),
			),
			mpb.AppendDecorators(
				decor.Counters(decor.CountersNoUnit(""), "%d / %d"),
			),
		)

		// Hide progresss bars when running in standard serial mode
		if !c.Parallel {
			bar.Abort(true)
		}

		bar.SetTotal(int64(len(target.Steps)), false)
		c.ExecuteSteps(target, target.Steps, bar)
		bar.EnableTriggerComplete()
	}
}

func (c *Client) ExecuteSteps(target targets.Target, steps []targets.Step, bar *mpb.Bar) {
	for _, step := range steps {
		switch step.Action {

		// Run commands on the remote server
		case "cmd":
			c.Printer.Default("Command:")
			c.Printer.Info(step.Command)

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

			c.Printer.Default("Result:")
			if err != nil {
				c.Printer.Error(stderr)
				c.Printer.Fatal(err)
			}

			c.Printer.Success(stdout)
			bar.Increment()

		// Copy a file to the remote server
		case "file":
			ctx := context.Background()
			c.Printer.Default("File:")
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
				c.Printer.Info(step.Source)
			}

			if step.Data != "" && step.Source == "" {
				c.Printer.Info(step.Data)
				if err := os.WriteFile(temp.Name()+"render.tmp", []byte(step.Data), 0644); err != nil {
					log.Fatalln(err)
				}
				file.Source = temp.Name() + "render.tmp"
			}

			_, stderr, err := ssh.RunCommand(target, "mkdir -p "+temp.Name()+path.Dir(file.Destination))
			if err != nil {
				c.Printer.Error(stderr)
				c.Printer.Fatal(err)
			}

			ssh.TransferFile(ctx, *target.User, target.Host, file, temp.Name())

			c.Printer.Default("Result:")

			_, stderr, err = ssh.RunCommand(target, "sudo mv "+temp.Name()+file.Destination+" "+file.Destination)
			if err != nil {
				c.Printer.Error(stderr)
				c.Printer.Fatal(err)
			}

			if step.Chown != "" {
				_, stderr, err := ssh.RunCommand(target, "sudo chown "+step.Chown+" "+step.Destination)
				if err != nil {
					c.Printer.Error(stderr)
					c.Printer.Fatal(err)
				}
			}

			_, stderr, err = ssh.RunCommand(target, "sudo rm -rf "+temp.Name())
			if err != nil {
				c.Printer.Error(stderr)
				c.Printer.Fatal(err)
			}

			c.Printer.Success(step.Destination)
			bar.Increment()

		// Parse a file with steps in it and run them
		case "task":
			// Recursively run a task
			config, err := targets.ParseHCL([]byte(step.Task), c.EvalContext)
			if err != nil {
				c.Printer.Fatal(err)
			}

			bar.SetTotal(int64(len(config.Steps))+bar.Current(), false)
			c.ExecuteSteps(target, config.Steps, bar)

		default:
			log.Fatalln("Unknown action", step.Action)
		}
	}
}
