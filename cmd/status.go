package cmd

import (
	"fmt"
	"path"
	"time"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

var prevCount int
var poll bool

func printServiceStatus(cluster, service string) bool {
	serviceObj, err := ecs.FindService(cluster, service)
	failOnError(err, fmt.Sprintf("Problem finding service: %s %s\n", cluster, service))
	fmt.Printf("Cluster:\t\t%s\n", cluster)
	fmt.Printf("Service:\t\t%s\n", service)
	fmt.Printf("Task Definition:\t%s\n", path.Base(*serviceObj.TaskDefinition))
	fmt.Println("Deployments:")
	for _, d := range serviceObj.Deployments {
		fmt.Printf("%s %s (%d Desired, %d Pending, %d Running) %v\n", path.Base(*d.TaskDefinition), *d.Status, *d.DesiredCount, *d.PendingCount, *d.RunningCount, *d.CreatedAt)
	}
	currCount := len(serviceObj.Deployments)
	hasChanged := prevCount != 1000 && currCount != prevCount
	prevCount = currCount
	return !hasChanged
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "View current cluster or service deployment status",
	Long:  "View current cluster or service deployment status",
	Run: func(cmd *cobra.Command, args []string) {
		cluster, service := ServiceChooser(args)
		prevCount = 1000
		printServiceStatus(cluster, service)
		for poll {
			time.Sleep(15 * time.Second)
			poll = printServiceStatus(cluster, service)
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	statusCmd.Flags().BoolVarP(&poll, "poll", "p", false, "Poll at a 15 second interval until a change in deployments occurs")

}
