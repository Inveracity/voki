package inline

import (
	"bytes"
	"html/template"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var TemplateFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
		{
			Name:             "map",
			Type:             cty.Map(cty.String),
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		in := args[0].AsString()
		out, err := renderTemplate(in, args[1].AsValueMap())
		return cty.StringVal(string(out)), err
	},
})

func renderTemplate(file string, data map[string]cty.Value) (string, error) {
	t, err := template.ParseFiles(file)
	if err != nil {
		return "", err
	}
	t.Option("missingkey=error")

	dataMap := make(map[string]interface{})
	for k, v := range data {
		dataMap[k] = v.AsString()
	}

	var wr bytes.Buffer
	if err := t.Execute(&wr, dataMap); err != nil {
		return "", err
	}

	return wr.String(), nil
}
