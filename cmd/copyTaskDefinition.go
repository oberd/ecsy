package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var commandOverride string
var newFamilyName string

// copyTaskDefinitionCmd represents the createTaskRevision command
var copyTaskDefinitionCmd = &cobra.Command{
	Use:   "copy-task-revision [cluster] [service] --image --task-definition-source --family-name --command-override",
	Short: "duplicate a task definition into a new revision with a different image",
	Long: `
example:
    ecsy copy-task-revision oberd-prod-mx care-guide-mx -s newest --family-name care-guide-mx-inbox-prod -c 'php artisan queue:work sqs-sns'
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if newFamilyName == "" {
			return errors.New("Please specify a family name for the new definition")
		}
		def, err := ecs.LocateTaskDef(args[0], args[1], taskDefinitionSource)
		if err != nil {
			return err
		}
		if imageUrl != "" {
			def, err = ecs.CreateNewTaskWithImage(def, imageUrl)
			if err != nil {
				return err
			}
		}
		newTaskDef, err := ecs.CopyTaskDefinition(def, newFamilyName, ecs.WithCommand(commandOverride))
		if err != nil {
			return err
		}
		fmt.Println("configured new task definition: ", newTaskDef)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(copyTaskDefinitionCmd)
	copyTaskDefinitionCmd.Flags().StringVarP(&taskDefinitionSource, "task-definition-source", "s", "newest", `locator for task definition ("newest" will use the newest in the family of the service, "current" will use the currently deployed)`)
	copyTaskDefinitionCmd.Flags().StringVarP(&imageUrl, "image", "i", "", "Image to update into the active container definition")
	copyTaskDefinitionCmd.Flags().StringVarP(&newFamilyName, "family-name", "f", "", "name of the new task definition family family")
	copyTaskDefinitionCmd.Flags().StringVarP(&commandOverride, "command-override", "c", "", "command for the new family")
}
