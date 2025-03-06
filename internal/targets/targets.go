package targets

import (
	"log"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclparse"
)

type Configuration struct {
	Targets []Target `hcl:"target,block"`
}

type Target struct {
	Name  string `hcl:"name,label"`
	User  string `hcl:"user"`
	Host  string `hcl:"host"`
	Steps []Step `hcl:"step,block"`
}

type Step struct {
	Action  string `hcl:"action,label"`
	Command string `hcl:"command"`
}

func Parse(specfile string) Configuration {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(specfile)

	if diags.HasErrors() {
		log.Fatal(diags)
	}

	var config Configuration
	confDiags := gohcl.DecodeBody(file.Body, nil, &config)

	if confDiags.HasErrors() {
		log.Fatal(confDiags)
	}

	return config
}
