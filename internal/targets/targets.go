package targets

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/inveracity/voki/internal/targets/inline"
	"github.com/zclconf/go-cty/cty/function"
)

// ParseHCL takes a byte array, from os.ReadFile
func ParseHCL(in []byte) (*Configuration, error) {
	// If there are variables in the target configuration, load those to add to ctx
	variables, err := LoadVars(in)
	if err != nil {
		log.Fatalln(err)
	}

	// Parse the rest of the target configuration
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
		Variables: variables,
	}

	var config Configuration
	diags = gohcl.DecodeBody(file.Body, ctx, &config)
	if diags.HasErrors() {
		log.Fatalln(diags)
	}

	return &config, nil
}
