package cli

import (
	"log"
	"os"
	"path/filepath"

	"github.com/inveracity/voki/internal/controller"
	"github.com/spf13/cobra"
)

// home returns the voki state directory for keys, pins and the peer database.
func home() string {
	if h := os.Getenv("VOKI_HOME"); h != "" {
		return h
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}
	return filepath.Join(dir, "voki")
}

type CmdServer struct{}

func (h *CmdServer) Command() *cobra.Command {
	var addr string
	cmd := &cobra.Command{
		Use:   "server",
		Short: "start the voki controller (agents connect here, run commands from this node)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Run(home(), addr)
		},
	}
	cmd.Flags().StringVarP(&addr, "addr", "a", ":7331", "address for agents to connect to")
	return cmd
}
