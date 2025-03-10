package targets

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty/function"
)

type Configuration struct {
	Targets []Target `hcl:"target,block"`
	Tasks   []Task   `hcl:"task,block"`
	Steps   []Step   `hcl:"step,block"`
}

type Target struct {
	Name  string `hcl:"name,label"`
	User  string `hcl:"user"`
	Host  string `hcl:"host"`
	Steps []Step `hcl:"step,block"`
}

type Step struct {
	Action      string   `hcl:"action,label"`
	Command     string   `hcl:"command,optional"`
	Task        string   `hcl:"task,optional"`
	Use         []string `hcl:"use,optional"`
	Source      string   `hcl:"source,optional"`
	Destination string   `hcl:"destination,optional"`
	Mode        string   `hcl:"mode,optional"`
}

type Task struct {
	Name  string `hcl:"task,label"`
	Steps []Step `hcl:"step,block"`
}

func ParseFile(filename string) (*Configuration, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filename)
	return parse(file, diags)
}

func ParseString(hcl []byte) (*Configuration, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(hcl, "stdin")
	return parse(file, diags)
}

func parse(file *hcl.File, diags hcl.Diagnostics) (*Configuration, error) {
	if diags.HasErrors() {
		log.Fatal(diags)
	}

	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"file":     FileFunc,
			"template": TemplateFunc,
		},
	}

	var config Configuration
	err := gohcl.DecodeBody(file.Body, ctx, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
