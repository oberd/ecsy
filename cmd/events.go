package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventsCmd = &cobra.Command{
	Use:   "events",
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
}

func init() {
	RootCmd.AddCommand(eventsCmd)
}
