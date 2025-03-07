package targets

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
		out, err := Template(in, args[1].AsValueMap())
		return cty.StringVal(string(out)), err
	},
})

func Template(file string, data map[string]cty.Value) (string, error) {
	t, err := template.ParseFiles(file)
	t.Option("missingkey=error")
	if err != nil {
		return "", err
	}

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
