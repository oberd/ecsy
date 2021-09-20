package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// refreshInstancesCmd represents the events command
var refreshInstanceCmd = &cobra.Command{
	Use:   "refresh-instances [cluster] [service]",
	Short: "Refresh instances (force deployment) for a service in a cluster",
	Long:  `Refresh instances (force deployment) for a service in a cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ecs.RefreshInstances(args[0], args[1])
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Incorrect number of arguments supplied! (%d / 2)", len(args))
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(refreshInstanceCmd)
}
