package targets

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hclparse"
)

type Configuration struct {
	Targets []Target `hcl:"targets,block"`
}

type Target struct {
	Name string `hcl:"name,optional"`
	User string `hcl:"user"`
	Host string `hcl:"host"`
	Cmd  string `hcl:"cmd,optional"`
}

func Parse() []Target {
	configfile, err := findConfig()
	if err != nil {
		log.Fatal(err)
	}
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(configfile)

	if diags.HasErrors() {
		log.Fatal(diags)
	}

	var config Configuration
	confDiags := gohcl.DecodeBody(file.Body, nil, &config)

	if confDiags.HasErrors() {
		log.Fatal(confDiags)
	}

	// PrintConfig(config.Targets)

	return config.Targets
}

func PrintConfig(config []Target) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', tabwriter.TabIndent)

	for _, target := range config {
		fmt.Fprintf(w, "%s\t%s\n", target.Name, target.Host)
	}
	w.Flush()
}

func findConfig() (string, error) {
	homedir, _ := os.UserHomeDir()
	currentDir := cwd()
	searchPaths := []string{
		currentDir + "/targets.hcl",
		homedir + "/.config/voki/targets.hcl",
		"/etc/voki/targets.hcl",
	}

	for _, spath := range searchPaths {
		if _, err := os.Stat(spath); err == nil {
			return spath, nil
		}
	}

	return "", errors.New("no targets.hcl found")
}

func cwd() string {
	ex, err := filepath.Abs("./")
	if err != nil {
		panic(err)
	}
	return ex
}
