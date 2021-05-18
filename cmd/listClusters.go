package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// listClustersCmd represents the listClusters command
var listClustersCmd = &cobra.Command{
	Use:   "list-clusters",
	Short: "lists clusters",
	Long:  `lists all clusters`,
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := ecs.ListClusters()
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
	RootCmd.AddCommand(listClustersCmd)
}
