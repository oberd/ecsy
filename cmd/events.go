package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventsCmd = &cobra.Command{
	Use:   "events [cluster] [service]",
	Short: "Show recent events for a service in a cluster",
	Long:  `Show recent events for a service in a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		svc, err := ecs.FindService(args[0], args[1])
		failOnError(err, "Error finding service")
		out := ""
		for _, event := range svc.Events {
			out = fmt.Sprintf("[%v] %s\n", *event.CreatedAt, *event.Message) + out
		}
		fmt.Print(out)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Incorrect number of arguments supplied! (%d / 2)", len(args))
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(eventsCmd)
}
