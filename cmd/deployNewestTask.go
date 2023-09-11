package cmd

import (
	"github.com/oberd/ecsy/ecs"

	"github.com/spf13/cobra"
)

//var taskArn string

// deployNewestTaskCmd represents the updateServiceTask command
var deployNewestTaskCmd = &cobra.Command{
	Use:   "deploy-newest-task [cluster] [service]",
	Short: "deploy newest task definition to a service",
	Long:  `by default, fast forwards to newest task definition`,
	Run: func(cmd *cobra.Command, args []string) {
		cluster, service := ServiceChooser(args)
		currentTask, err := ecs.GetCurrentTaskDefinition(cluster, service)
		failOnError(err, "finding current definition")
		newestTask, err := ecs.FindNewestDefinition(*currentTask.Family)
		failOnError(err, "getting newest definition")
		if newestTask.TaskDefinitionArn == currentTask.TaskDefinitionArn {
			return
		}
		_, err = ecs.DeployTaskToService(cluster, service, newestTask)
		failOnError(err, "deploying to service")
	},
}

func init() {
	RootCmd.AddCommand(deployNewestTaskCmd)
	//deployNewestTaskCmd.Flags().StringVarP(&taskArn, "taskArn", "t", "", "Help message for toggle")
}
