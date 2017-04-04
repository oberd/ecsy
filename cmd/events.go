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

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Show recent events for a service in a cluster",
	Long:  `Show recent events for a service in a cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		svc, err := ecs.FindService(args[0], args[1])
		failOnError(err, "Error finding service")
		out := ""
		for _, event := range svc.Events {
			out = fmt.Sprintf("[%v] %s\n", *event.CreatedAt, *event.Message) + out
		}
		fmt.Print(out)
	},
}

func init() {
	RootCmd.AddCommand(eventsCmd)
}
