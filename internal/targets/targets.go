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
	Name string `hcl:"name,label"`
	User string `hcl:"user"`
	Host string `hcl:"host"`
	Cmd  string `hcl:"cmd,optional"`
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
