package targets

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/inveracity/voki/internal/targets/inline"
	"github.com/zclconf/go-cty/cty/function"
)

type Configuration struct {
	Targets []Target `hcl:"target,block"`
	Tasks   []Task   `hcl:"task,block"`
	Steps   []Step   `hcl:"step,block"`
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

// ParseHCL takes a byte array, from os.ReadFile
func ParseHCL(in []byte) (*Configuration, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(in, "stdin")
	if diags.HasErrors() {
		log.Fatal(diags)
	}

	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"file":     inline.FileFunc,
			"template": inline.TemplateFunc,
		},
	}

	var config Configuration
	err := gohcl.DecodeBody(file.Body, ctx, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
