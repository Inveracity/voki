package cli

import (
	"github.com/inveracity/voki/internal/client"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
)

type App struct {
	Client client.Client
}

func New() *App {
	return &App{}
}

func (a *App) Run() error {
	rootCmd := &cobra.Command{
		Use:   "voki",
		Short: "voki cli",
		Long: dedent.Dedent(`
			usage: voki --help
			`,
		),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			InitializeEnv(cmd)
			a.Client = *client.New()
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand((&CmdPing{Client: &a.Client}).Command())
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true

	return rootCmd.Execute()
}
