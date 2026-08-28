package cli

import (
	"time"

	"github.com/inveracity/voki/internal/agent"
	"github.com/spf13/cobra"
)

type CmdAgent struct{}

func (h *CmdAgent) Command() *cobra.Command {
	var (
		controller string
		retry      time.Duration
	)
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "start the voki agent (run this on machines you manage)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agent.Run(home(), controller, retry)
		},
	}
	cmd.Flags().StringVar(&controller, "controller", "", "controller address host:port")
	cmd.Flags().DurationVar(&retry, "retry", 5*time.Second, "how long to wait before retrying a failed or pending connection")
	_ = cmd.MarkFlagRequired("controller")
	return cmd
}
