package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale [cluster] [service] [desired-count]",
	Short: "Set the number of desired instances of a service",
	Long:  "Set the number of desired instances of a service",
	Run: func(cmd *cobra.Command, args []string) {
		cluster, service := ServiceChooser(args)
		if len(args) < 3 {
			failOnError(fmt.Errorf("Not enough arguments"), "")
		}
		desiredCount, err := strconv.Atoi(args[2])
		failOnError(err, "Not able to parse integer")
		svc, err := ecs.FindService(cluster, service)
		failOnError(err, "")
		if int(*svc.DesiredCount) == desiredCount {
			fmt.Printf("Service already set to scale (%d)\n", desiredCount)
			os.Exit(0)
		}
		_, err = ecs.ScaleService(cluster, service, desiredCount)
		failOnError(err, "Unable to set service scale")
		fmt.Println("Successfully scaled service")
		printServiceStatus(cluster, service)
	},
}

func init() {
	RootCmd.AddCommand(scaleCmd)
}
