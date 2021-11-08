package cmd

import (
	"strings"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

var createPostDeploymentTask = &cobra.Command{
	Use:   "create-post-deployment-task",
	Short: "Creates an events rule that runs an ecs task with [command] after a service reaches steady state",
	Long: `Easily create a post-deployment task event
Example:
    ecsy create-post-deployment-task \
        --source-cluster=mountain-qa \
        --source-service=mountain-dashboards-qa \
        --target-cluster=general-prod \
        --target-service=slack-notifier \
        'notify-slack "Mountain QA Deployed"'
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		input := &ecs.CreatePostDeploymentTaskInput{}
		if err := cmd.Flags().Parse(args); err != nil {
			return err
		}
		input.SourceCluster = cmd.Flag("source-cluster").Value.String()
		input.SourceService = cmd.Flag("source-service").Value.String()
		input.TargetCluster = cmd.Flag("target-cluster").Value.String()
		input.TargetService = cmd.Flag("target-service").Value.String()
		if len(args) == 0 {
			return cmd.Usage()
		}
		input.Command = strings.Join(args[0:], " ")
		return ecs.CreatePostDeploymentTask(input)
	},
}

func init() {
	RootCmd.AddCommand(createPostDeploymentTask)
	createPostDeploymentTask.Flags().StringP("source-cluster", "", "", "cluster on which to listen for steady state events")
	createPostDeploymentTask.MarkFlagRequired("source-cluster")
	createPostDeploymentTask.Flags().StringP("source-service", "", "", "service source to listen for steady state events")
	createPostDeploymentTask.MarkFlagRequired("source-service")
	createPostDeploymentTask.Flags().StringP("target-cluster", "", "", "cluster on which to run the target task")
	createPostDeploymentTask.MarkFlagRequired("target-cluster")
	createPostDeploymentTask.Flags().StringP("target-service", "", "", "service to target for the task")
	createPostDeploymentTask.MarkFlagRequired("target-service")
}
