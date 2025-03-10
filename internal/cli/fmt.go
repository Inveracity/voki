package cli

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/inveracity/voki/internal/client"
)

var (
	write     bool
	recursive bool
)

type CmdFmt struct {
	Client *client.Client
}

func (h *CmdFmt) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "fmt",
		Short:         "format a hcl file",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				log.Fatalln("expected 1 argument")
			}

			f := client.FormatCommand{Paths: args, WriteFile: write, Recursive: recursive, WriteStdout: !write}
			f.Fmt()
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVarP(&write, "write", "w", false, "write file")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse through sub directories")
	return cmd
}
