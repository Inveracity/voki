package targets

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/vault-client-go"
	"github.com/zclconf/go-cty/cty"
)

type VaultConfig struct {
	VaultToken string
}

type Vault struct {
	Mountpath string `hcl:"mountpath"`
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

	// TODO: if the vault block is not in the hcl file body, this crashes with a runtime error index out of range [0]
	blockVault, blockDiags := vaultBody.Blocks[0].Body.Content(vaultBlockSchema)
	if blockDiags.HasErrors() {
		return nil, err
	}

	var c = &Vault{}
	if attr, exists := blockVault.Attributes["mountpath"]; exists {
		decodeDiags := gohcl.DecodeExpression(attr.Expr, nil, &c.Mountpath)
		if decodeDiags.HasErrors() {
			log.Fatalf("decode mountpath attr error: %v", decodeDiags)
		}
		bleg, err := v.getSecret(c.Mountpath)
		fmt.Println(bleg)
		return bleg, err
	}

	return map[string]cty.Value{}, nil
}

func (v *VaultConfig) getSecret(mountpath string) (map[string]cty.Value, error) {
	ctx := context.Background()

	client, err := vault.New(
		vault.WithAddress("http://127.0.0.1:8200"),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.SetToken(v.VaultToken); err != nil {
		log.Fatal(err)
	}

	s, err := client.Secrets.KvV2Read(ctx, mountpath, vault.WithMountPath("secret"))
	if err != nil {
		log.Fatal("vault ", err)
	}

	variables := make(map[string]cty.Value)
	for name, val := range s.Data.Data {
		variables[name] = cty.StringVal(fmt.Sprintf("%v", val))
	}
	return variables, nil
}
