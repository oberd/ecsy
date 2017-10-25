package ecs

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

type serviceKey string

func newServiceKey(cluster, service string) serviceKey {
	return serviceKey(fmt.Sprintf("%s_%s", cluster, service))
}

var _ecs *ecs.ECS
var _ec2 *ec2.EC2
var _cloudWatch *cloudwatchlogs.CloudWatchLogs
var _clusterArns map[string]string

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

func assertCloudWatch(region string) *cloudwatchlogs.CloudWatchLogs {
	if _cloudWatch == nil {
		_cloudWatch = cloudwatchlogs.New(session.New(&aws.Config{Region: aws.String(region)}))
	}
	return _cloudWatch
}

func assertClusterMap() (map[string]string, error) {
	if _clusterArns == nil {
		svc := assertECS()
		clusters, err := svc.ListClusters(&ecs.ListClustersInput{})
		if err != nil {
			return nil, err
		}
		_clusterArns = make(map[string]string, len(clusters.ClusterArns))
		for _, s := range clusters.ClusterArns {
			parts := strings.Split(*s, "/")
			simpleName := parts[len(parts)-1]
			_clusterArns[simpleName] = *s
		}
	}
	return _clusterArns, nil
}

// GetClusterNames returns a slice of strings representing cluster names
func GetClusterNames() ([]string, error) {
	clusterMap, err := assertClusterMap()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(clusterMap))
	i := 0
	for name := range clusterMap {
		out[i] = name
		i++
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
	return nil, fmt.Errorf("Error finding service with name %s in cluster %s, %d results found.", service, cluster, len(result.Services))
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

// CreateNewTaskWithEnvironment registers a new task, based on the passed task,
// but with new environment.
func CreateNewTaskWithEnvironment(existingTask *ecs.TaskDefinition, env []*ecs.KeyValuePair) (*ecs.TaskDefinition, error) {
	var found bool
	for _, def := range existingTask.ContainerDefinitions {
		if *def.Essential {
			def.SetEnvironment(env)
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("Error finding essential container, does the task %s have a container marked as essential?\n", existingTask.GoString())
	}
	svc := assertECS()
	input := &ecs.RegisterTaskDefinitionInput{}
	input.SetContainerDefinitions(existingTask.ContainerDefinitions)
	input.SetFamily(*existingTask.Family)
	input.SetVolumes(existingTask.Volumes)
	newTaskDef, err := svc.RegisterTaskDefinition(input)
	if err != nil {
		return nil, err
	}
	return newTaskDef.TaskDefinition, nil
}

// FindService finds a service struct by name
func FindService(cluster, service string) (*ecs.Service, error) {
	svc := assertECS()
	result, err := svc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(cluster),
		Services: []*string{aws.String(service)},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Services) == 1 {
		return result.Services[0], nil
	}
	return nil, fmt.Errorf("Did not find one (%d) services matching name %s in %s cluster, unable to continue.", len(result.Services), service, cluster)
}

// ListServices lists services for a cluster by plain name
func ListServices(cluster string) ([]string, error) {
	svc := assertECS()
	params := &ecs.ListServicesInput{Cluster: aws.String(cluster)}
	result := make([]*string, 0)
	err := svc.ListServicesPages(params, func(services *ecs.ListServicesOutput, lastPage bool) bool {
		result = append(result, services.ServiceArns...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, len(result))
	for i, s := range result {
		out[i] = path.Base(*s)
	}
	return out, nil
}

// DeployTaskToService deploys a given task definition to a cluster/service
func DeployTaskToService(cluster, service string, task *ecs.TaskDefinition) (*ecs.Service, error) {
	input := &ecs.UpdateServiceInput{}
	input.SetCluster(cluster)
	input.SetService(service)
	input.SetTaskDefinition(*task.TaskDefinitionArn)
	svc := assertECS()
	output, err := svc.UpdateService(input)
	if err != nil {
		return nil, err
	}
	return output.Service, nil
}

// ScaleService sets the desired count of a service
func ScaleService(cluster, service string, desiredCount int) (*ecs.Service, error) {
	input := &ecs.UpdateServiceInput{}
	input.SetCluster(cluster)
	input.SetService(service)
	input.SetDesiredCount(int64(desiredCount))
	svc := assertECS()
	output, err := svc.UpdateService(input)
	if err != nil {
		return nil, err
	}
	return output.Service, nil
}

// KeyPairsToString takes a list of key pairs... and prints them
// into a multiline block
func KeyPairsToString(kv []*ecs.KeyValuePair) string {
	out := ""
	for _, envVar := range kv {
		out += fmt.Sprintf("%s=%s\n", *envVar.Name, *envVar.Value)
	}
	return out
}

// StringToKeyPairs takes a string, and turns it back into
// an array of key pairs
func StringToKeyPairs(input string) ([]*ecs.KeyValuePair, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	output := make([]*ecs.KeyValuePair, 0)
	i := 0
	for scanner.Scan() {
		i++
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Problem parsing line %d: %s", i, line)
		}
		newPair := &ecs.KeyValuePair{}
		newPair.SetName(parts[0])
		newPair.SetValue(parts[1])
		output = append(output, newPair)
	}
	return output, nil
}

// BuildConsoleURLForService builds the console url for a service
func BuildConsoleURLForService(cluster, service string) string {
	region := os.Getenv("AWS_REGION")
	return fmt.Sprintf("https://%s.console.aws.amazon.com/ecs/home?region=%s#/clusters/%s/services/%s", region, region, cluster, service)
}

// GetLogs returns cloudwatch logs for a specific set of log streams
// matching a pattern within a log group and region
func GetLogs(cluster, service string) error {
	def, err := GetCurrentTaskDefinition(cluster, service)
	if err != nil {
		return fmt.Errorf("Problem getting task definition: %v", err)
	}
	logConfig := def.ContainerDefinitions[0].LogConfiguration
	if *logConfig.LogDriver != "awslogs" {
		fmt.Print(fmt.Errorf("logs command requires awslogs driver, found %v", def.ContainerDefinitions[0].LogConfiguration))
		os.Exit(1)
	}
	group := logConfig.Options["awslogs-group"]
	prefix := logConfig.Options["awslogs-stream-prefix"]
	region := *logConfig.Options["awslogs-region"]
	svc := assertCloudWatch(region)
	pageNum := 0
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: group,
		OrderBy:      aws.String("LastEventTime"),
		Descending:   aws.Bool(true),
	}
	var eventErr error
	err = svc.DescribeLogStreamsPages(params, func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
		pageNum++
		for _, stream := range page.LogStreams {
			name := stream.LogStreamName
			if strings.HasPrefix(*name, *prefix) {
				eventErr := ListLogEvents(*group, *name, region)
				if eventErr != nil {
					return false
				}
			}
		}
		return pageNum <= 3
	})
	if err != nil {
		return fmt.Errorf("Problem listing log streams: %v", err)
	}
	if eventErr != nil {
		return fmt.Errorf("Problem listing log streams: %v", eventErr)
	}
	return nil
}

// ListLogEvents accepts a single group and stream name
// and prints the cloudwatch logs to stdout
func ListLogEvents(group, name, region string) error {
	svc := assertCloudWatch(region)
	eventParams := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(name),
	}
	eventPage := 0
	eventErr := svc.GetLogEventsPages(eventParams, func(events *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
		eventPage++
		for _, event := range events.Events {
			fmt.Println(*event.Message)
		}
		return eventPage <= 3
	})
	if eventErr != nil {
		return eventErr
	}
	return nil
}
