package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// listServicesCmd represents the listServices command
var listServicesCmd = &cobra.Command{
	Use:   "list-services",
	Short: "list services in a cluster",
	Long:  `list services in a cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide an argument for [cluster]")
		}
		list, err := ecs.ListServices(args[0])
		if err != nil {
			return err
		}
		for _, svc := range list {
			fmt.Println(svc)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(listServicesCmd)
}
