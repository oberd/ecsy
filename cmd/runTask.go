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
	"strings"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

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
		output, err := ecs.RunTaskWithCommand(cluster, service, command)
		if err != nil {
			log.Fatalf("%v\n", err)
			return
		}
		parts := strings.Split(*output.Tasks[0].TaskArn, "/")
		taskID := parts[len(parts)-1]
		detailsLink := fmt.Sprintf(
			"https://us-west-2.console.aws.amazon.com/ecs/home?region=us-west-2#/clusters/%s/tasks/%s/details",
			cluster,
			taskID,
		)
		fmt.Printf("For Details, Visit:\n%s\n", detailsLink)
	},
}

func init() {
	RootCmd.AddCommand(runTaskCmd)
}
