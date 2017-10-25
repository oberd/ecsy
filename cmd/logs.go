package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs [cluster] [service]",
	Short: "Show recent logs for a service in a cluster (must be cloudwatch based)",
	Long:  `Show recent logs for a service in a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ecs.GetLogs(args[0], args[1])
		failOnError(err, "")
		// out := ""
		// for _, event := range svc.Events {
		// 	out = fmt.Sprintf("[%v] %s\n", *event.CreatedAt, *event.Message) + out
		// }
		// fmt.Print(out)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Incorrect number of arguments supplied! (%d / 2)", len(args))
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(logsCmd)
}
