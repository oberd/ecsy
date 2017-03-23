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

// sshCmd represents the exec command
var sshCmd = &cobra.Command{
	Use:   "ssh [cluster] [service]",
	Short: "Secure Shell into one of the service container instances' EC2 host machines",
	Long: `
Lets say you have a service running a few container instances on various hosts.
Lets say you also need to debug some file configuration running in the container,
or you maybe need to check some files on disk of the Host, you can shell right
in there through this command using a menu system rather than figuring it out
yourself! Yay. I guess.
`,
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
	RootCmd.AddCommand(sshCmd)
}
