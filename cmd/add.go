package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/config"
	"github.com/spf13/cobra"
)

// register-keyCmd represents the register-key command
var addCmd = &cobra.Command{
	Use:   "add [cluster-name] [key-file]",
	Short: "Associates a .pem SSH key with a cluster, allowing SSH into EC2 instances",
	Long:  `You only have to do this once per cluster, ecsy will store the stuff in config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file := config.GetYAMLConfig()
		keys, err := file.ReadKeys()
		if err != nil {
			return err
		}
		clusterName := config.ClusterName(args[0])
		keys[clusterName] = config.FilePath(args[1])
		err = file.SaveKeys(keys)
		if err != nil {
			return err
		}
		fmt.Printf("Successfully wrote key pair: %s %s\n", args[0], args[1])
		return nil
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Not enough arguemnts supplied! %v", args)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
	addCmd.ArgAliases = []string{"cluster", "pemFile"}
}
