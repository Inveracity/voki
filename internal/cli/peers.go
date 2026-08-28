package cli

import (
	"fmt"
	"log"
	"strconv"

	"github.com/inveracity/voki/internal/controller"
	"github.com/inveracity/voki/internal/peers"
	"github.com/spf13/cobra"
)

type CmdPeers struct{}

func (h *CmdPeers) Command() *cobra.Command {
	root := &cobra.Command{
		Use:   "peers",
		Short: "manage agents allowed to connect to this controller",
	}

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "list agents and their status",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadPeers()
			if err != nil {
				return err
			}
			if len(db.Peers) == 0 {
				log.Println("no agents. run 'voki agent --controller <this-host>' on a managed machine")
				return nil
			}
			for _, p := range db.Peers {
				log.Printf("%d  %-10s  %s  %s", p.ID, p.Status, p.Fingerprint[:16], p.Name)
			}
			return nil
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "accept <id>",
		Short: "accept a pending agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setPeerStatus(args[0], "accepted")
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "reject <id>",
		Short: "reject a pending or accepted agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setPeerStatus(args[0], "rejected")
		},
	})

	return root
}

func loadPeers() (*peers.DB, error) {
	return peers.Load(controller.PeersPath(home()))
}

func setPeerStatus(idArg, status string) error {
	id, err := strconv.Atoi(idArg)
	if err != nil {
		return fmt.Errorf("peer id must be a number: %w", err)
	}

	db, err := loadPeers()
	if err != nil {
		return err
	}

	if err := db.SetStatus(id, peers.Status(status)); err != nil {
		return err
	}

	log.Printf("agent %d %s", id, status)
	return nil
}
