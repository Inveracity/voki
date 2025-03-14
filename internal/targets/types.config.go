package targets

import "github.com/hashicorp/hcl/v2"

type Configuration struct {
	Targets []Target `hcl:"target,block"`
	Tasks   []Task   `hcl:"task,block"`
	Steps   []Step   `hcl:"step,block"`
	Remain  hcl.Body `hcl:",remain"`
}

type Target struct {
	Name  string  `hcl:"name,label"`
	User  *string `hcl:"user"`
	Host  string  `hcl:"host"`
	Steps []Step  `hcl:"step,block"`
}

type Step struct {
	Action      string   `hcl:"action,label"`
	Command     string   `hcl:"command,optional"`
	Task        string   `hcl:"task,optional"`
	Use         []string `hcl:"use,optional"`
	Source      string   `hcl:"source,optional"`
	Data        string   `hcl:"data,optional"`
	Destination string   `hcl:"destination,optional"`
	Mode        string   `hcl:"mode,optional"`
	Sudo        bool     `hcl:"sudo,optional"`
	Shell       string   `hcl:"shell,optional"`
	Chown       string   `hcl:"chown,optional"`
}

type Task struct {
	Name  string `hcl:"task,label"`
	Steps []Step `hcl:"step,block"`
}
