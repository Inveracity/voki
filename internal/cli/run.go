package cli

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/inveracity/voki/internal/client"
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
