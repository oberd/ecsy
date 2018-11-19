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
	"strings"

	"github.com/oberd/ecsy/config"
	"github.com/oberd/ecsy/ecs"
	"github.com/oberd/ecsy/ssh"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [cluster]",
	Short: "Run an ssh command on all the servers in a cluster",
	Long:  `Run an ssh command on all the servers in a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		clusterKey := config.GetClusterKey(cluster)
		commandParts := args[1:]
		command := strings.Join(commandParts, " ")
		instances, err := ecs.GetContainerInstances(cluster, "")
		if err != nil {
			log.Fatalf("Problem retrieving servers: %v", err)
		}
		for _, instance := range instances {
			client := ssh.NewClient(ssh.ClientConfiguration{
				Host:           instance,
				User:           "ec2-user",
				PrivateKeyFile: clusterKey,
			})
			output, err := client.Run(command, false)
			if err != nil {
				log.Fatalf("Problem running command: %v", err)
			}
			fmt.Printf("%v\n", output)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
