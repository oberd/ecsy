package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// refreshCmd
var refreshCmd = &cobra.Command{
	Use:   "refresh [cluster] [service]",
	Short: "force a new deployment (same number of tasks, but refresh them)",
	Long:  "in some cases the connections have died for your service (or maybe there is a boot configuration file that needs updating), you can use this command to refresh the runtime containers of a task without updating its task definition or code",
	RunE: func(cmd *cobra.Command, args []string) error {
		cluster, service := ServiceChooser(args)
		if len(args) < 2 {
			return fmt.Errorf("Not enough arguments")
		}
		fmt.Printf("forcing new deployment for %s/%s", cluster, service)
		return ecs.CreateRefreshDeployment(cluster, service)
	},
}

func init() {
	RootCmd.AddCommand(refreshCmd)
}
