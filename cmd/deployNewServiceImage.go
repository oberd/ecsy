
package cmd

import (
	"fmt"
    "github.com/oberd/ecsy/ecs"

    "github.com/spf13/cobra"
)

// deployNewServiceImageCmd represents the deployNewServiceImage command
var deployNewServiceImageCmd = &cobra.Command{
	Use:   "deploy [cluster] [service] [image]",
	Short: "deploys a new image to a cluster service",
	Long: `creates a new task revision cloning the current service one, and updates the service`,
	Run: func(cmd *cobra.Command, args []string) {
	    cluster, service := ServiceChooser(args)
        def, err := ecs.GetCurrentTaskDefinition(cluster, service)
        failOnError(err, "getting task definition")
        if len(args) != 3 {
            failOnError(fmt.Errorf("please specify an image"), "bad arguments")
        }
        newTask, err := ecs.FindNewestDefinition(*def.Family)
        if err != nil || ecs.EssentialImage(newTask) != args[2] {
            newTask, err = ecs.CreateNewTaskWithImage(def, args[2])
            fmt.Printf("created task definition with image %s\n", args[2])
            failOnError(err, "create new task def")
        }
        svc, err := ecs.DeployTaskToService(cluster, service, newTask)
        failOnError(err, "updating service task")
        fmt.Printf("updated service %s with task definition %s (deploying to %d containers)", *svc.ServiceArn, *newTask.TaskDefinitionArn, *svc.DesiredCount)
	},
}

func init() {
	RootCmd.AddCommand(deployNewServiceImageCmd)
}
