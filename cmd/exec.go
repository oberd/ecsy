// Copyright © 2017 Nathan Bleigh <nathan@oberd.com>
//

package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/config"
	"github.com/oberd/ecsy/ecs"
	"github.com/oberd/ecsy/ssh"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec [cluster] [service]",
	Short: "Execute a command in a running container task, on a remote EC2 host",
	Long:  `Lends you the ability to dive right into a running container.`,
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		service := args[1]
		err := ecs.ValidateCluster(cluster)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		res, err := ecs.GetContainerInstances(cluster, service)
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		clusterKey := config.GetClusterKey(cluster)
		choice := StringChooser(res, "Please choose a server")
		client := ssh.NewClient(ssh.ClientConfiguration{
			Host:           choice,
			User:           "ec2-user",
			PrivateKeyFile: clusterKey,
		})
		err = client.Connect()
		if err != nil {
			fmt.Printf("%v", err)
		}
		fmt.Printf("Finished\n")
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Not enough arguemnts supplied! %v", args)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// execCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// execCmd.Flags().StringP("toggle", "t", false, "Help message for toggle")

}
