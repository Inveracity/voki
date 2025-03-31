package inline

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/vault-client-go"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

var VaultFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.Map(cty.String)),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		in := args[0].AsString()
		out, err := getSecret(in)
		r, err := gocty.ToCtyValue(out, cty.Map(cty.String))
		return r, err
	},
})

func getSecret(mountpath string) (map[string]interface{}, error) {
	ctx := context.Background()

	// prepare a client with the given base address
	client, err := vault.New(
		vault.WithAddress("http://127.0.0.1:8200"),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	// authenticate with a root token (insecure)
	if err := client.SetToken("123456"); err != nil {
		log.Fatal(err)
	}

	// read the secret
	s, err := client.Secrets.KvV2Read(ctx, mountpath, vault.WithMountPath("secret"))
	if err != nil {
		log.Fatal(err)
	}
	return s.Data.Data, nil
}
