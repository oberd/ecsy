package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

var imageUrl string

// createTaskRevisionCmd represents the createTaskRevision command
var createTaskRevisionCmd = &cobra.Command{
	Use:   "create-task-revision [cluster] [service] --image --task-definition-source",
	Short: "duplicate a task definition into a new revision with a different image",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		def, err := ecs.LocateTaskDef(args[0], args[1], taskDefinitionSource)
		if err != nil {
			return err
		}
		newTask, err := ecs.CreateNewTaskWithImage(def, imageUrl)
		if err != nil {
			return err
		}
		fmt.Printf("%s", *newTask.TaskDefinitionArn)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(createTaskRevisionCmd)
	createTaskRevisionCmd.Flags().StringVarP(&taskDefinitionSource, "task-definition-source", "s", "newest", `locator for task definition ("newest" will use the newest in the family of the service, "current" will use the currently deployed)`)
	createTaskRevisionCmd.Flags().StringVarP(&imageUrl, "image", "i", "", "Image to update into the active container definition")
}
