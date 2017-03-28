package ecs

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

var _ecs *ecs.ECS
var _ec2 *ec2.EC2

func assertECS() *ecs.ECS {
	if _ecs == nil {
		_ecs = ecs.New(session.New(&aws.Config{Region: aws.String("us-west-2")}))
	}
	return _ecs
}

func assertEC2() *ec2.EC2 {
	if _ec2 == nil {
		_ec2 = ec2.New(session.New(&aws.Config{Region: aws.String("us-west-2")}))
	}
	return _ec2
}

// GetClusterNames returns a slice of strings representing cluster names
func GetClusterNames() ([]string, error) {
	svc := assertECS()
	clusters, err := svc.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}
	out := make([]string, len(clusters.ClusterArns))
	for i, s := range clusters.ClusterArns {
		parts := strings.Split(*s, "/")
		out[i] = parts[len(parts)-1]
	}
	return out, nil
}

// ValidateCluster returns an error if the cluster is not found
func ValidateCluster(cluster string) error {
	names, err := GetClusterNames()
	if err != nil {
		return err
	}
	for _, name := range names {
		if name == cluster {
			return nil
		}
	}
	return fmt.Errorf("Cluster not found: %s", cluster)
}

// GetDeployedEssentialContainer gets the currently deployed "essential" container
// definition of a cluster service
func GetDeployedEssentialContainer(cluster, service string) (*ecs.ContainerDefinition, error) {
	err := ValidateCluster(cluster)
	if err != nil {
		return nil, err
	}
	task, err := GetCurrentTaskDefinition(cluster, service)
	if err != nil {
		return nil, err
	}
	container, err := GetEssentialContainer(task)
	if err != nil {
		return nil, err
	}
	return container, nil
}

// GetCurrentTaskDefinition returns a service's current task definition
func GetCurrentTaskDefinition(cluster, service string) (*ecs.TaskDefinition, error) {
	svc := assertECS()
	input := &ecs.DescribeServicesInput{}
	input.SetCluster(cluster)
	input.SetServices([]*string{aws.String(service)})
	result, err := svc.DescribeServices(input)
	if err != nil {
		return nil, err
	}
	if len(result.Services) == 1 {
		service := result.Services[0]
		describeTaskInput := &ecs.DescribeTaskDefinitionInput{}
		describeTaskInput.SetTaskDefinition(*service.TaskDefinition)
		findTaskOutput, err := svc.DescribeTaskDefinition(describeTaskInput)
		if err != nil {
			return nil, err
		}
		return findTaskOutput.TaskDefinition, nil
	}
	return nil, fmt.Errorf("Error finding service with name %s in cluster %s, %d results found.", cluster, service, len(result.Services))
}

// GetEssentialContainer returns the essential container
func GetEssentialContainer(task *ecs.TaskDefinition) (*ecs.ContainerDefinition, error) {
	for _, def := range task.ContainerDefinitions {
		if *def.Essential {
			return def, nil
		}
	}
	return nil, fmt.Errorf("Error finding essential container, does the task %s have a container marked as essential?\n", task.GoString())
}

// GetContainerInstances returns the container instances of a cluster
func GetContainerInstances(cluster string, service string) ([]string, error) {
	svc := assertECS()
	input := &ecs.ListTasksInput{}
	input.SetCluster(cluster)
	input.SetServiceName(service)
	result, err := svc.ListTasks(input)
	if err != nil {
		return nil, err
	}
	input2 := &ecs.DescribeTasksInput{}
	input2.SetCluster(cluster)
	input2.SetTasks(result.TaskArns)
	result2, err := svc.DescribeTasks(input2)
	if err != nil {
		return nil, err
	}
	instances := make([]*string, len(result2.Tasks))
	for i, t := range result2.Tasks {
		instances[i] = t.ContainerInstanceArn
	}
	instanceInput := &ecs.DescribeContainerInstancesInput{}
	instanceInput.SetCluster(cluster)
	instanceInput.SetContainerInstances(instances)
	result3, err := svc.DescribeContainerInstances(instanceInput)
	if err != nil {
		return nil, err
	}
	ec2svc := assertEC2()
	iinput := &ec2.DescribeInstancesInput{}
	ec2instances := make([]*string, len(result3.ContainerInstances))
	for i, ci := range result3.ContainerInstances {
		ec2instances[i] = ci.Ec2InstanceId
	}
	iinput.SetInstanceIds(ec2instances)
	result4, err := ec2svc.DescribeInstances(iinput)
	if err != nil {
		return nil, err
	}
	dnsNames := make([]string, len(result4.Reservations))
	for i, r := range result4.Reservations {
		dnsNames[i] = *r.Instances[0].PublicDnsName
	}
	return dnsNames, nil
}
