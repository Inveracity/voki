package client

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/inveracity/voki/internal/local"
	"github.com/inveracity/voki/internal/targets"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

// Executor runs steps against one target, either locally or over the voki protocol.
type Executor interface {
	Exec(command string) (stdout, stderr string, err error)
	Upload(step targets.Step) error
	Close() error
}

type LocalExecutor struct{}

func (LocalExecutor) Exec(command string) (string, string, error) {
	return local.RunCommand(targets.Step{Shell: "bash", Command: command})
}

func (LocalExecutor) Upload(step targets.Step) error {
	targets.DiffFiles(step.Destination, step.Source)
	return local.CopyFile(step.Source, step.Destination)
}

func (LocalExecutor) Close() error { return nil }

// NewExecutor picks the executor for a target host.
// A Dialer returns a connected remote executor.
type Dialer func(host string) (Executor, error)

func (c *Client) executor(target targets.Target, dial Dialer) (Executor, error) {
	if isLocal(target.Host) {
		return LocalExecutor{}, nil
	}
	return dial(target.Host)
}

func isLocal(host string) bool {
	host = strings.ToLower(host)
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func (c *Client) Run(hcl string, dial Dialer) {
	config, err := targets.ParseHCL([]byte(hcl), c.EvalContext)
	if err != nil {
		log.Fatalln(err)
	}

	for _, target := range config.Targets {
		c.Printer.Title(target.Name)

		var bar *mpb.Bar
		if c.Parallel {
			bar = c.Bar.AddBar(0,
				mpb.BarOptional(mpb.BarWidth(40), true),
				mpb.PrependDecorators(
					decor.Name(fmt.Sprintf("%-27s", target.Name)),
				),
				mpb.AppendDecorators(
					decor.Counters(decor.CountersNoUnit(""), "%d / %d"),
				),
			)
		}

		if bar != nil {
			bar.SetTotal(int64(len(target.Steps)), false)
		}

		ex, err := c.executor(target, dial)
		if err != nil {
			c.Printer.Fatal(err)
		}

		c.ExecuteSteps(ex, target.Steps, bar)
		_ = ex.Close()

		if bar != nil {
			bar.EnableTriggerComplete()
		}
	}
}

func (c *Client) ExecuteSteps(ex Executor, steps []targets.Step, bar *mpb.Bar) {
	for _, step := range steps {
		// SkipStep checks if the step should be skipped based on the provided steps
		// If no steps are provided, all steps will be executed
		if c.Steps != nil && len(*c.Steps) > 0 && !slices.Contains(*c.Steps, step.Name) {
			continue
		}

		switch step.Action {

		// Run commands on the target
		case "cmd":
			c.Printer.Default(fmt.Sprintf("Command: %s", step.Name))
			c.Printer.Info(step.Command)

			stdout, stderr, err := ex.Exec(wrapShell(step))

			c.Printer.Default("Result:")
			if err != nil {
				c.Printer.Error(stderr)
				c.Printer.Fatal(err)
			}

			if stdout != "" {
				c.Printer.Success(stdout)
			}

			if stderr != "" {
				c.Printer.Error(stderr)
			}

			if bar != nil {
				bar.Increment()
			}

		// Copy a file to the target
		case "file":
			c.Printer.Default("File:")

			if step.Source != "" {
				c.Printer.Info(step.Source)
			}

			if step.Data != "" && step.Source == "" {
				temp, err := os.CreateTemp("", ".voki-*")
				if err != nil {
					log.Fatalln(err)
				}
				defer os.Remove(temp.Name())
				if err := os.WriteFile(temp.Name(), []byte(step.Data), 0644); err != nil {
					log.Fatalln(err)
				}
				step.Source = temp.Name()
				c.Printer.Info(step.Data)
			}

			c.Printer.Default("Result:")
			if err := ex.Upload(step); err != nil {
				c.Printer.Fatal(err)
			}
			c.Printer.Success(step.Destination)

			if bar != nil {
				bar.Increment()
			}

		// Parse a file with steps in it and run them
		case "task":
			config, err := targets.ParseHCL([]byte(step.Task), c.EvalContext)
			if err != nil {
				c.Printer.Fatal(err)
			}

			if bar != nil {
				bar.SetTotal(int64(len(config.Steps))+bar.Current(), false)
			}
			c.ExecuteSteps(ex, config.Steps, bar)

		default:
			log.Fatalln("Unknown action", step.Action)
		}
	}
}

// wrapShell builds the final command: [sudo] <shell> -c '<command>'.
func wrapShell(step targets.Step) string {
	shell := step.Shell
	if shell == "" {
		shell = "bash"
	}
	cmd := fmt.Sprintf("%s -c '%s'", shell, step.Command)
	if step.Sudo {
		cmd = "sudo " + cmd
	}
	return cmd
}
