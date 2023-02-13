package cmd

import (
	"fmt"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

var memoryFlag int64 = 0
var memoryReservationFlag int64 = 0

// copyTaskDefinitionCmd represents the createTaskRevision command
var setMemoryCmd = &cobra.Command{
	Use:   "set-memory [cluster] [service] --memory [int megabytes] --memory-reservation [int megabytes]",
	Short: "set memory properties on the essential container of a service",
	Long: `
set memory properties on the essential container of a service
    see https://aws.amazon.com/blogs/containers/how-amazon-ecs-manages-cpu-and-memory-resources/
    for more information about memory reservations
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		def, err := ecs.LocateTaskDef(args[0], args[1], "newest")
		if err != nil {
			return err
		}
		var memoryParam *int64
		var memoryReservationParam *int64
		if memoryFlag != -1 {
			memoryParam = &memoryFlag
		}
		if memoryReservationFlag != -1 {
			memoryReservationParam = &memoryReservationFlag
		}
		newTaskDef, err := ecs.CreateNewTaskWithMemory(def, memoryParam, memoryReservationParam)
		if err != nil {
			return err
		}
		fmt.Println("configured new task definition: ", newTaskDef)
		service, err := ecs.DeployTaskToService(args[0], args[1], newTaskDef)
		if err != nil {
			return err
		}
		fmt.Printf("deployed new memory to %s %s\n", *service.ClusterArn, *service.ServiceName)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(setMemoryCmd)
	setMemoryCmd.Flags().Int64VarP(&memoryFlag, "memory", "m", -1, `"memory" for a task`)
	setMemoryCmd.Flags().Int64VarP(&memoryReservationFlag, "memory-reservation", "r", -1, `"memoryReservation" for a task`)
}
