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
