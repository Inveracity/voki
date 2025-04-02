package targets

import (
	"os"
	"strings"

	"github.com/zclconf/go-cty/cty"
)

func GetEnv() map[string]cty.Value {
	env := map[string]cty.Value{}
	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) == 2 {
			env[kv[0]] = cty.StringVal(kv[1])
		}
	}
	return env
}
