package targets

import (
	"log"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/zclconf/go-cty/cty/function"
)

type Configuration struct {
	Targets []Target      `hcl:"target,block"`
	Import  []ImportBlock `hcl:"import,block"`
	Tasks   []Task        `hcl:"task,block"`
}

type ImportBlock struct {
	File string `hcl:"file"`
}

type Target struct {
	Name  string      `hcl:"name,label"`
	User  string      `hcl:"user"`
	Host  string      `hcl:"host"`
	Steps []Step      `hcl:"step,block"`
	Apply *ApplyBlock `hcl:"apply,block"`
}

type Step struct {
	Action  string `hcl:"action,label"`
	Command string `hcl:"command"`
}

type Task struct {
	Name  string `hcl:"task,label"`
	Steps []Step `hcl:"step,block"`
}

type ApplyBlock struct {
	Use []string `hcl:"use"`
}

func Parse(filename string) (*Configuration, error) {
	var config Configuration

	parser := hclparse.NewParser()

	file, diags := parser.ParseHCLFile(filename)

	if diags.HasErrors() {
		log.Fatal(diags)
	}

	ctx := &hcl.EvalContext{
		Functions: map[string]function.Function{
			"file":     FileFunc,
			"template": TemplateFunc,
		},
	}

	err := gohcl.DecodeBody(file.Body, ctx, &config)
	if err != nil {
		return nil, err
	}

	// Handle import if specified
	if config.Import != nil {
		for _, importBlock := range config.Import {
			importedConfig, err := Parse(importBlock.File)
			if err != nil {
				return nil, err
			}

			if importedConfig.Tasks != nil {
				config.Tasks = append(config.Tasks, importedConfig.Tasks...)
			}
		}
	}

	return &config, nil
}
