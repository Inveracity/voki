package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Read environment variables with the THING_ prefix
func InitializeEnv(cmd *cobra.Command) {
	v := viper.New()
	v.SetEnvPrefix("VOKI")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	v.SetConfigType("hcl")
	v.SetConfigFile("voki-config.hcl")
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		log.Printf("Error reading config file: %s", err)
	}

	bindFlags(cmd, v)
}

// Bind each cobra flag to its associated viper configuration environment variable
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				log.Panic(err)
			}
		}
	})
}
