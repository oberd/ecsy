// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"log"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// updateAgentCmd represents the updateAgent command
var updateAgentCmd = &cobra.Command{
	Use:   "update-agent [cluster-name]",
	Short: "Update the Container Instance Agents",
	Long:  `This will grab a list of the container instances and upgrade their agents (performs a cluster wide upgrade)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Cluster not specified")
		}
		cluster := args[0]
		instances, err := ecs.GetClusterInstances(cluster)
		if err != nil {
			log.Fatalf("error retrieving instances: %v", err)
		}
		count := len(instances)
		fmt.Printf("Found %d instances\n", count)
		for i := 0; i < count; i++ {
			fmt.Printf("Upgrading agent %d of %d\n", i, count)
			err := ecs.UpdateAgent(cluster, instances[i])
			if err != nil {
				return err
			}
		}
		fmt.Printf("Finished upgrading %d instance agents\n", count)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(updateAgentCmd)
}
