package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// describeCmd represents the logs command
var describeCmd = &cobra.Command{
	Use:   "describe [cluster] [service]",
	Short: "Show current task configuration for service",
	Long:  `Show current task configuration for service`,
	Run: func(cmd *cobra.Command, args []string) {
		def, err := ecs.GetCurrentTaskDefinition(args[0], args[1])
		failOnError(err, "Error finding service")
		fmt.Printf("%v\n", def)
		service, err := ecs.FindService(args[0], args[1])
		failOnError(err, "Error finding service")
		fmt.Println("")
		fmt.Println("AWS Console URLs")
		fmt.Println("================")
		fmt.Println("")
		taskURLs, err := ecs.PrintTaskURLs(args[0], service)
		failOnError(err, "error getting task urls")
		for _, uri := range taskURLs {
			fmt.Println(uri)
			fmt.Println("")
		}
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Incorrect number of arguments supplied! (%d / 2)", len(args))
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(describeCmd)
}
