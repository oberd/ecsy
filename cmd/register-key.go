// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/config"
	"github.com/spf13/cobra"
)

// register-keyCmd represents the register-key command
var registerKeyCmd = &cobra.Command{
	Use:   "add [cluster-name] [key-file]",
	Short: "Associates a .pem SSH key with a cluster, allowing SSH into EC2 instances",
	Long:  `You only have to do this once per cluster, ecs-commander will store the stuff in config.`,
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
	RootCmd.AddCommand(registerKeyCmd)

	// Here you will define your flags and configuration settings.

	registerKeyCmd.ArgAliases = []string{"cluster", "pemFile"}

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// register-keyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// register-keyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
