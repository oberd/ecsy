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
	"os"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env [command]",
	Short: "Used to manage environment variables of service task definitions",
	Long: `Getting and setting, then deploying new environment variables in ECS services
can be a little annoying due to the cumbersome web ui.  Do it in one fell swoop with
ecsy env set ENV_VAR=value`,
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [cluster-name] [service-name]",
	Short: "List environment variables for an ECS service's deployed task definition",
	Long:  `List environment variables for an ECS service's deployed task definition`,
	Run: func(cmd *cobra.Command, args []string) {
		container, err := ecs.GetDeployedEssentialContainer(args[0], args[1])
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		for _, envVar := range container.Environment {
			fmt.Printf("%s=%s\n", *envVar.Name, *envVar.Value)
		}
	},
	PreRunE: ValidateArgumentsCount2,
}

func init() {
	RootCmd.AddCommand(envCmd)
	envCmd.AddCommand(getCmd)
}
