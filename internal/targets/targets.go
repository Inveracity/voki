package targets

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ParseHCL takes a byte array, from os.ReadFile
func ParseHCL(in []byte, evalcontext *hcl.EvalContext) (*Configuration, error) {

	// Parse the rest of the target configuration
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(in, "stdin")
	if diags.HasErrors() {
		log.Fatal(diags)
	}

	var config Configuration
	diags = gohcl.DecodeBody(file.Body, evalcontext, &config)
	if diags.HasErrors() {
		log.Fatalln(diags)
	}

	return &config, nil
}
