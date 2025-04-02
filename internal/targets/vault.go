package targets

import (
	"context"
	"fmt"
	"log"
	"maps"
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
	Alias     string `hcl:"alias"`
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
		{Name: "alias"},
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

		if attr, exists := blockVault.Attributes["alias"]; exists {
			decodeDiags := gohcl.DecodeExpression(attr.Expr, nil, &c.Alias)
			if decodeDiags.HasErrors() {
				log.Fatalf("decode path attr error: %v", decodeDiags)
			}
		}

		if c.Mountpath == "" || c.Path == "" || c.Alias == "" {
			return nil, fmt.Errorf("mountpath, path and alias are required")
		}

		res, err := v.getSecret(c.Mountpath, c.Path)
		if err != nil {
			return nil, err
		}

		aliased := map[string]cty.Value{c.Alias: cty.ObjectVal(res)}

		maps.Copy(secrets, aliased)
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
