package client

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/pborman/indent"

	"github.com/inveracity/voki/internal/targets"
)

func (c *Client) Run(specfile string) {
	config, err := targets.Parse(specfile)
	if err != nil {
		log.Fatalln(err)
	}
	w := indent.New(os.Stdout, "   ")

	for _, target := range config.Targets {
		fmt.Println("==== " + target.Name + " ====\n")

		// Add imported steps from a task file
		if target.Apply != nil {
			for _, taskname := range target.Apply.Use {
				task, err := findTask(config.Tasks, taskname)
				if err != nil {
					log.Fatalln(err)
				}
				target.Steps = append(target.Steps, task.Steps...)
			}
		}

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

func findTask(tasks []targets.Task, name string) (*targets.Task, error) {
	for _, task := range tasks {
		if task.Name == name {
			return &task, nil
		}
	}
	return nil, fmt.Errorf("task %s not found", name)
}
