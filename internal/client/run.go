package client

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/pborman/indent"

	"github.com/inveracity/voki/internal/targets"
)

func (c *Client) Run(specfile string) {
	config := targets.Parse(specfile)
	w := indent.New(os.Stdout, "   ")

	for _, target := range config.Targets {
		fmt.Println("==== " + target.Name + " ====\n")

		for idx, step := range target.Steps {
			if step.Action == "cmd" {
				fmt.Println("Command", idx+1)
				fmt.Fprintln(w, color.BlueString(step.Command))
				result := TestConnection(target, step.Command)
				fmt.Println("Result:")
				fmt.Fprintln(w, color.GreenString(result))
			}
		}
	}
}
