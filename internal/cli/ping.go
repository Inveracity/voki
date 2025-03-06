package cli

import (
	"github.com/spf13/cobra"

	"github.com/inveracity/voki/internal/client"
)

type CmdPing struct {
	Client *client.Client
}

func (h *CmdPing) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ping",
		Short:         "Test if server is reachable",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				h.Client.Ping("")
				return nil
			}
			h.Client.Ping(args[0])
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
