package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"

	"github.com/spf13/cobra"
)

// scheduleTaskCmd represents the scheduleTask command
var scheduleTaskCmd = &cobra.Command{
	Use:   "schedule-task cluster service taskSuffix schedule command",
	Short: "Creates a scheduled task with a command override",
	Long: `Easily create a cron, or other scheduled task using the command line.
Example:
    ecsy schedule-task mountain-qa mountain-dashboards-qa generate-snapshots 'rate(1 day)' 'npm run generate-snapshots'
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 5 {
			return fmt.Errorf("Not enough arguments")
		}
		if len(args) > 5 {
			return fmt.Errorf("Too many arguments, make sure you quote your command")
		}
		cluster, service := ServiceChooser(args)
		suffix := args[2]
		schedule := args[3]
		command := args[4]
		err := ecs.CreateScheduledTask(cluster, service, suffix, schedule, command)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(scheduleTaskCmd)
}
