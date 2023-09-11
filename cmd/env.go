package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"

	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/oberd/ecsy/ecs"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env [command]",
	Short: "Used to manage environment variables of service task definitions",
	Long: `Getting and setting, then deploying new environment variables in ECS services
can be a little annoying due to the cumbersome web ui.  Do it in one fell swoop with
ecsy env set or ecsy env edit`,
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [cluster-name] [service-name]",
	Short: "List environment variables for an ECS service's deployed task definition",
	Long:  `List environment variables for an ECS service's deployed task definition`,
	Run: func(cmd *cobra.Command, args []string) {
		container, err := ecs.GetDeployedEssentialContainer(args[0], args[1])
		failOnError(err, "")
		fmt.Print(ecs.KeyPairsToString(container.Environment))
	},
	PreRunE: Validate2ArgumentsCount,
}

// setCmd sets a variable
var setCmd = &cobra.Command{
	Use:   "set [cluster-name] [service-name] [env_var_name] [env_var_value]",
	Short: "Set an environment variable for an ECS service's deployed task definition",
	Long:  `Set an environment variable for an ECS service's deployed task definition`,
	Run: func(cmd *cobra.Command, args []string) {
		container, err := ecs.GetDeployedEssentialContainer(args[0], args[1])
		failOnError(err, "")
		varName := args[2]
		varValue := args[3]
		var found bool
		for _, val := range container.Environment {
			if *val.Name == varName {
				val.SetValue(varValue)
				found = true
			}
		}
		if !found {
			container.Environment = append(container.Environment, &awsecs.KeyValuePair{Name: aws.String(varName), Value: aws.String(varValue)})
		}
		deployEnv(args[0], args[1], container.Environment)
	},
	PreRunE: Validate4ArgumentsCount,
}

// editCmd allows you to edit env vars of an active service's
// task definition
var editCmd = &cobra.Command{
	Use:   "edit [cluster-name] [service-name]",
	Short: "Interactively define environment for a task, and deploy it to the service",
	Long:  `Interactively define environment for a task, and deploy it to the service`,
	Run: func(cmd *cobra.Command, args []string) {
		cluster := args[0]
		service := args[1]
		primary, err := ecs.GetDeployedEssentialContainer(cluster, service)
		if err != nil {
			fmt.Printf("Error finding essential container:\n%v\n", err)
			os.Exit(1)
		}
		original := ecs.KeyPairsToString(primary.Environment)
		original = strings.TrimSpace(original)
		edited, err := EditStringBlock(original)
		edited = strings.TrimSpace(edited)
		if err != nil {
			fmt.Printf("Error editing environment: %v\n", err)
			os.Exit(1)
		}
		if original == edited {
			fmt.Println("No changes made to environment.  Nothing to do!")
		} else {
			newKeyPairs, err := ecs.StringToKeyPairs(edited)
			if err != nil {
				fmt.Printf("Problem parsing new environment: %v\n", err)
				os.Exit(1)
			}
			confirm := `
Do you want to save the following environment to the task?

%s

This will also update service "%s" in "%s" to a new task definition.`
			if !AskForConfirmation(fmt.Sprintf(confirm, ecs.KeyPairsToString(newKeyPairs), cluster, service)) {
				return
			}
			deployEnv(args[0], args[1], newKeyPairs)
		}
	},
	PreRunE: Validate2ArgumentsCount,
}

var findCmd = &cobra.Command{
	Use:   "find [env_var_name] [env_var_value]",
	Short: "Find all services that have an environment variable with the given name and value",
	Long:  `Find all services that have an environment variable with the given name and value`,
	Run: func(cmd *cobra.Command, args []string) {
		varName := args[0]
		varValue := args[1]
		services, err := ecs.FindServicesWithEnvVar(varName, varValue)
		if err != nil {
			fmt.Printf("Problem finding services: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d services with %s=%s:\n", len(services), varName, varValue)
		for _, service := range services {
			fmt.Printf("%s %s\n", *service.ClusterArn, *service.ServiceName)
		}
	},
	PreRunE: Validate2ArgumentsCount,
}

func init() {
	RootCmd.AddCommand(envCmd)
	envCmd.AddCommand(getCmd)
	envCmd.AddCommand(editCmd)
	envCmd.AddCommand(setCmd)
	envCmd.AddCommand(findCmd)
}

func deployEnv(cluster, service string, newKeyPairs []*awsecs.KeyValuePair) {
	fmt.Println("Getting current task definition...")
	task, err := ecs.GetCurrentTaskDefinition(cluster, service)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Creating new task based on %s:%d, with new environment\n", *task.Family, *task.Revision)
	newTask, err := ecs.CreateNewTaskWithEnvironment(task, newKeyPairs)
	if err != nil {
		fmt.Printf("Problem creating new task: %v\n", err)
		os.Exit(1)
	}
	serviceStruct, err := ecs.DeployTaskToService(cluster, service, newTask)
	if err != nil {
		fmt.Printf("Problem deploying task: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("\nSuccessfully deployed new task definition")
	fmt.Println("=========================================")
	fmt.Printf("Cluster: %s\n", cluster)
	fmt.Printf("Service: %s\n", service)
	fmt.Printf("Task Definition: %s:%d\n", *newTask.Family, *newTask.Revision)
	fmt.Printf("Desired Count: %d\n", *serviceStruct.DesiredCount)
	fmt.Printf("To view deployment status, you can visit:\n%s\n", ecs.BuildConsoleURLForService(cluster, service)+"/deployments")
}
