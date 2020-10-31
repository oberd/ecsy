// Copyright © 2018 Nathan Bleigh <nathan@oberd.com>
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
	"os"
	"time"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

var runTaskWait = false
var taskDefinitionSource = "newest"

// runTaskCmd represents the runTask command
var runTaskCmd = &cobra.Command{
	Use:   "run-task [cluster] [service] '[command override]'",
	Short: "Run an individual task into an ECS cluster",
	Long: `Typically, you will want to run this in the case
that you have a one-off command to run in a cluster, and
don't want to SSH in. Make sure you enclose your command
in single quotes.

For example, to run an artisan or cron command

Example:

ecsy run-task medamine indexer-worker 'bin/snapshot --configuration assets/medamine_configurations/knee_replacement_revision_medamine.json'
`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		cluster, service := ServiceChooser(args)
		if len(args) < 3 {
			failOnError(fmt.Errorf("Not enough arguments"), "")
		}
		command := args[2]
		def, err := ecs.LocateTaskDef(cluster, service, taskDefinitionSource)
		output, err := ecs.RunTaskWithCommand(cluster, def, command)
		if err != nil {
			log.Fatalf("%v\n", err)
			return
		}
		if len(output.Failures) > 0 {
			for _, failure := range output.Failures {
				log.Fatalf("Received failure from AWS:\n%v", failure.String())
			}
			return
		}
		taskID := ecs.GetTaskIDFromArn(*output.Tasks[0].TaskArn)
		detailsLink := fmt.Sprintf(
			"https://us-west-2.console.aws.amazon.com/ecs/home?region=us-west-2#/clusters/%s/tasks/%s/details",
			cluster,
			taskID,
		)
		fmt.Printf("=> Created Task: %s\n", detailsLink)
		if runTaskWait {
			var exitCode *int64
			fmt.Printf("==> Waiting for task to complete...\n")
			for exitCode == nil {
				task, err := ecs.GetTask(cluster, *output.Tasks[0].TaskArn)
				if err != nil {
					log.Fatalln("could not find task", err)
				}
				time.Sleep(time.Second * 2)
				exitCode, err = ecs.TaskExitCode(task)
			}
			if err != nil {
				log.Fatalf("%v\n", err)
				return
			}
			fmt.Printf("==> Retrieving Container Log Output\n")
			err = ecs.GetTaskLogs(def, taskID)
			if err != nil {
				log.Printf("Unable to read logs: %v\n", err)
			}
			fmt.Printf("==> End Container Log Output\n")
			if *exitCode > 0 {
				log.Printf("==> Received error code from container: %v", *exitCode)
			} else {
				fmt.Printf("=> Tasks Completed Successfully\n")
			}
			os.Exit(int(*exitCode))
		}
	},
}

func init() {
	RootCmd.AddCommand(runTaskCmd)
	runTaskCmd.Flags().StringVarP(&taskDefinitionSource, "task-definition-source", "s", "newest", `locator for task definition ("newest" will use the newest in the family of the service, "current" will use the currently deployed)`)
	runTaskCmd.Flags().BoolVarP(&runTaskWait, "wait", "w", false, "Wait for completion of task")
}
