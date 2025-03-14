package targets

import (
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

type Vars struct {
	//Target hcl.Body `hcl:"target,optional"`
	Remain hcl.Body `hcl:",remain"`
}

func LoadVars(in []byte) (map[string]cty.Value, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(in, "stdin")
	if diags.HasErrors() {
		log.Fatal(diags)
	}

	var vars Vars
	err := gohcl.DecodeBody(file.Body, nil, &vars)
	if err != nil {
		return nil, err
	}

	return varsToCtyMap(&vars)
}

func varsToCtyMap(vars *Vars) (map[string]cty.Value, error) {
	if vars == nil {
		return nil, nil
	}

	// Grab the attributes and ignore diagnostic errors to partially load the contents of the hcl file
	attrs, _ := vars.Remain.JustAttributes()

	// Convert attributes into cty values
	variables := make(map[string]cty.Value)
	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags
		}
		variables[name] = val
	}
	return variables, nil
}
