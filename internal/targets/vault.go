package targets

import (
	"context"
	"fmt"
	"log"
	"maps"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/vault-client-go"
	"github.com/zclconf/go-cty/cty"
)

type VaultConfig struct {
	Token string
	Addr  string
}

type Vault struct {
	Mountpath string `hcl:"mountpath"`
	Path      string `hcl:"path"`
}

var vaultSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "vault",
		},
	},
}

var vaultBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "mountpath"},
		{Name: "path"},
	},
}

func (v *VaultConfig) LoadVault(in []byte) (map[string]cty.Value, error) {
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

	vaultBody, diags := file.Body.Content(vaultSchema)

	secrets := map[string]cty.Value{}
	for _, block := range vaultBody.Blocks {
		blockVault, blockDiags := block.Body.Content(vaultBlockSchema)

		if blockDiags.HasErrors() {
			return nil, err
		}

		var c = &Vault{}
		if attr, exists := blockVault.Attributes["mountpath"]; exists {
			decodeDiags := gohcl.DecodeExpression(attr.Expr, nil, &c.Mountpath)
			if decodeDiags.HasErrors() {
				log.Fatalf("decode mountpath attr error: %v", decodeDiags)
			}
		}

		if attr, exists := blockVault.Attributes["path"]; exists {
			decodeDiags := gohcl.DecodeExpression(attr.Expr, nil, &c.Path)
			if decodeDiags.HasErrors() {
				log.Fatalf("decode path attr error: %v", decodeDiags)
			}
		}

		if c.Mountpath == "" || c.Path == "" {
			return nil, fmt.Errorf("mountpath and path are required")
		}

		res, err := v.getSecret(c.Mountpath, c.Path)
		if err != nil {
			return nil, err
		}

		nesting := make(map[string]cty.Value)
		keys := strings.Split(c.Path, "/")
		fmt.Println(keys)
		for i, p := range keys {
			fmt.Println(i, len(keys)-1)
			if i == 0 {
				nesting = map[string]cty.Value{p: cty.ObjectVal(res)}
			} else {
				fmt.Println("-----", p, nesting)
				nesting[p] = cty.ObjectVal(nesting)
			}
		}

		fmt.Println(nesting)

		//values := map[string]cty.Value{p: cty.ObjectVal(res)}

		maps.Copy(secrets, nesting)
	}

	return secrets, nil
}

func (v *VaultConfig) getSecret(mountpath, path string) (map[string]cty.Value, error) {
	ctx := context.Background()

	client, err := vault.New(
		vault.WithAddress(v.Addr),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.SetToken(v.Token); err != nil {
		log.Fatal(err)
	}

	s, err := client.Secrets.KvV2Read(ctx, path, vault.WithMountPath(mountpath))
	if err != nil {
		log.Fatal("vault ", err)
	}

	variables := make(map[string]cty.Value)
	for name, val := range s.Data.Data {
		variables[name] = cty.StringVal(fmt.Sprintf("%v", val))
	}
	return variables, nil
}
