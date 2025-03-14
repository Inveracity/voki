package cli

import (
	"log"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/inveracity/voki/internal/client"
	"github.com/inveracity/voki/internal/targets"
	"github.com/inveracity/voki/internal/targets/inline"
)

var (
	user string
)

type CmdRun struct {
	Client *client.Client
}

func (h *CmdRun) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "run",
		Short:         "run a voki specification",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				log.Fatalln("expected 1 argument")
			}
			content, err := os.ReadFile(args[0])
			if err != nil {
				log.Fatalln(err)
			} // If there are variables in the target configuration, load those to add to ctx
			variables, err := targets.LoadVars(content)
			if err != nil {
				log.Fatalln(err)
			}

			h.Client.EvalContext = &hcl.EvalContext{
				Functions: map[string]function.Function{
					"file":     inline.FileFunc,
					"template": inline.TemplateFunc,
				},
				Variables: variables,
			}

			h.Client.Run(string(content), user)
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(&user, "user", "u", "", "user")
	viper.BindPFlag("user", cmd.PersistentFlags().Lookup("user"))
	return cmd
}
