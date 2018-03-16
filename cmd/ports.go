package cmd

import (
	"fmt"

	"github.com/docker/machine/libmachine/log"
	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// portsCmd represents the ports command
var portsCmd = &cobra.Command{
	Use:   "ports [cluster]",
	Short: "List out exposed service ports for creating new services",
	Long:  `When you are creating a new exposed port on the servers, nice to see what is already taken!`,
	Run: func(cmd *cobra.Command, args []string) {
		services, err := ecs.ListServices(args[0])
		if err != nil {
			log.Errorf("error fetching services: %v", err)
			return
		}
		for _, service := range services {
			ports := make([]int64, 0)
			taskDef, err := ecs.GetCurrentTaskDefinition(args[0], service)
			if err == nil {
				for _, container := range taskDef.ContainerDefinitions {
					for _, port := range container.PortMappings {
						ports = append(ports, *port.HostPort)
					}
				}
				fmt.Printf("%s %v\n", service, ports)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(portsCmd)
}
