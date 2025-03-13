package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/inveracity/voki/internal/version"
)

type CmdVersion struct {
}

func (h *CmdVersion) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "show version",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version.Version)
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
