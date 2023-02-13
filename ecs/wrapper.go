package ecs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/iam"
)

type serviceKey string

func newServiceKey(cluster, service string) serviceKey {
	return serviceKey(fmt.Sprintf("%s_%s", cluster, service))
}

var _ecs *ecs.ECS
var _ec2 *ec2.EC2
var _iam *iam.IAM
var _cloudWatch *cloudwatchlogs.CloudWatchLogs
var _cloudWatchEvents *cloudwatchevents.CloudWatchEvents
var _clusterArns map[string]string

func getServiceConfiguration() *aws.Config {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}
	return &aws.Config{Region: aws.String(region)}
}

func assertECS() *ecs.ECS {
	if _ecs == nil {
		_ecs = ecs.New(session.New(getServiceConfiguration()))
	}
	return _ecs
}

func assertEC2() *ec2.EC2 {
	if _ec2 == nil {
		_ec2 = ec2.New(session.New(getServiceConfiguration()))
	}
	return _ec2
}

func assertCloudWatch(region string) *cloudwatchlogs.CloudWatchLogs {
	if _cloudWatch == nil {
		_cloudWatch = cloudwatchlogs.New(session.New(getServiceConfiguration()))
	}
	return _cloudWatch
}

func assertCloudWatchEvents() *cloudwatchevents.CloudWatchEvents {
	if _cloudWatchEvents == nil {
		_cloudWatchEvents = cloudwatchevents.New(session.New(getServiceConfiguration()))
	}
	return _cloudWatchEvents
}
func assertIAM() *iam.IAM {
	if _iam == nil {
		_iam = iam.New(session.New(getServiceConfiguration()))
	}
	return _iam
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
	return nil, fmt.Errorf("error finding service with name %s in cluster %s, %d results found", service, cluster, len(result.Services))
}

func GetNewestTaskDefinition(cluster, service string) (*ecs.TaskDefinition, error) {
	def, err := GetCurrentTaskDefinition(cluster, service)
	if err != nil {
		return nil, err
	}
	svc := assertECS()
	input := &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: def.Family,
		Sort:         aws.String("DESC"),
		MaxResults:   aws.Int64(1),
	}
	results, err := svc.ListTaskDefinitions(input)
	if err != nil {
		return nil, err
	}
	if len(results.TaskDefinitionArns) == 0 {
		return nil, fmt.Errorf("task definitions not found for family %s", *def.Family)
	}
	return GetTaskDefinition(*results.TaskDefinitionArns[0])
}

// GetTaskDefinition returns a task definition by arn
func GetTaskDefinition(arn string) (*ecs.TaskDefinition, error) {
	svc := assertECS()
	output, err := svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(arn),
	})
	if err != nil {
		return nil, err
	}
	return output.TaskDefinition, nil
}

// GetEssentialContainer returns the essential container
func GetEssentialContainer(task *ecs.TaskDefinition) (*ecs.ContainerDefinition, error) {
	for _, def := range task.ContainerDefinitions {
		if *def.Essential {
			return def, nil
		}
	}
	return nil, fmt.Errorf("error finding essential container, does the task %s have a container marked as essential", task.GoString())
}

// GetClusterInstances returns the container instances of a cluster
// by id
func GetClusterInstances(cluster string) ([]string, error) {
	svc := assertECS()
	input := &ecs.ListContainerInstancesInput{
		Cluster: &cluster,
	}
	pageNum := 0
	output := make([]string, 0)
	err := svc.ListContainerInstancesPages(input,
		func(page *ecs.ListContainerInstancesOutput, lastPage bool) bool {
			for _, arn := range page.ContainerInstanceArns {
				output = append(output, *arn)
			}
			pageNum++
			return true
		})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// UpdateAgent will update the agents for all instances in a cluster
func UpdateAgent(cluster, containerInstanceARN string) error {
	svc := assertECS()
	input := &ecs.UpdateContainerAgentInput{
		Cluster:           &cluster,
		ContainerInstance: &containerInstanceARN,
	}
	_, err := svc.UpdateContainerAgent(input)
	if err != nil {
		return err
	}
	return nil
}

// GetContainerInstances returns the container instances of a cluster
func GetContainerInstances(cluster string, service string) ([]string, error) {
	svc := assertECS()
	input := &ecs.ListTasksInput{}
	input.SetCluster(cluster)
	if service != "" {
		input.SetServiceName(service)
	}
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
	return updateEssential(existingTask, func(container *ecs.ContainerDefinition) {
		container.SetEnvironment(env)
	})
}

// CreateNewTaskWithMemory registers a new task definition, overwriting memory
// params
func CreateNewTaskWithMemory(existingTask *ecs.TaskDefinition, memory *int64, memoryReservation *int64) (*ecs.TaskDefinition, error) {
	return updateEssential(existingTask, func(container *ecs.ContainerDefinition) {
		if memory != nil {
			container.SetMemory(*memory)
		}
		if memoryReservation != nil {
			container.SetMemoryReservation(*memoryReservation)
		}
	})
}

// CreateNewTaskWithImage registers a new task, based on the passed task,
// but with new image.
func CreateNewTaskWithImage(existingTask *ecs.TaskDefinition, imageURL string) (*ecs.TaskDefinition, error) {
	if imageURL == "" {
		return nil, fmt.Errorf("invalid imageUrl: empty")
	}
	return updateEssential(existingTask, func(container *ecs.ContainerDefinition) {
		container.SetImage(imageURL)
	})
}

func WithCommand(command string) func(definition *ecs.ContainerDefinition) {
	return func(def *ecs.ContainerDefinition) {
		parts, err := parseCommandOverride(command)
		if err != nil {
			log.Fatalln(err)
		}
		if len(parts) == 0 {
			return
		}
		def.SetCommand(mapCommandStrPointer(parts))
	}
}

func WithName(name string) func(definition *ecs.ContainerDefinition) {
	return func(def *ecs.ContainerDefinition) {
		def.SetName(name)
	}
}

// CopyTaskDefinitionWithCommand registers a new task, based on the passed task,
// but with new command.
func CopyTaskDefinition(existingTask *ecs.TaskDefinition, toFamilyName string, modifiers ...func(*ecs.ContainerDefinition)) (*ecs.TaskDefinition, error) {
	svc := assertECS()
	newTask := &ecs.TaskDefinition{
		ContainerDefinitions: existingTask.ContainerDefinitions,
		Family:               aws.String(toFamilyName),
	}
	essential := findEssential(existingTask)
	if essential == nil {
		return nil, errors.New("essential container not found")
	}
	WithName(toFamilyName)(essential)
	for _, modifier := range modifiers {
		modifier(essential)
	}
	output, err := svc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:               newTask.Family,
		ContainerDefinitions: newTask.ContainerDefinitions,
	})
	if err != nil {
		return nil, errors.Wrap(err, "task registration")
	}
	return output.TaskDefinition, nil
}

func updateEssential(existingTask *ecs.TaskDefinition, configure func(definition *ecs.ContainerDefinition)) (*ecs.TaskDefinition, error) {
	essential := findEssential(existingTask)
	if essential == nil {
		return nil, fmt.Errorf("error finding essential container, does the task %s have a container marked as essential", existingTask.GoString())
	}
	configure(essential)
	return saveTaskDef(existingTask)
}

func findEssential(task *ecs.TaskDefinition) *ecs.ContainerDefinition {
	for _, def := range task.ContainerDefinitions {
		if *def.Essential {
			return def
		}
	}
	return nil
}

func saveTaskDef(existingTask *ecs.TaskDefinition) (*ecs.TaskDefinition, error) {
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

// EssentialImage returns the essential image of a task def
func EssentialImage(task *ecs.TaskDefinition) string {
	essential := findEssential(task)
	if essential == nil {
		return ""
	}
	return *essential.Image
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
	return nil, fmt.Errorf("did not find one (%d) services matching name %s in %s cluster, unable to continue", len(result.Services), service, cluster)
}

func ListClusters() ([]string, error) {
	svc := assertECS()
	params := &ecs.ListClustersInput{}
	result := make([]*string, 0)
	err := svc.ListClustersPages(params, func(clusters *ecs.ListClustersOutput, lastPage bool) bool {
		result = append(result, clusters.ClusterArns...)
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

// FindNewestDefinition finds the most recent task definition of the service's
// task family
func FindNewestDefinition(familyPrefix string) (*ecs.TaskDefinition, error) {
	svc := assertECS()
	result, err := svc.ListTaskDefinitions(&ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(familyPrefix),
		Sort:         aws.String("DESC"),
		MaxResults:   aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.TaskDefinitionArns) == 0 {
		return nil, fmt.Errorf("could not find any task definitions for family %s", familyPrefix)
	}
	taskResult, err := svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{TaskDefinition: result.TaskDefinitionArns[0]})
	if err != nil {
		return nil, err
	}
	return taskResult.TaskDefinition, nil
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

// GetAllTasksByDefinition gets the tasks that have run recently for a
// cluster and service
func GetAllTasksByDefinition(cluster string, def *ecs.TaskDefinition, status string) ([]*ecs.Task, error) {
	runningTasks, err := GetAllTasksByDefinitionStatus(cluster, def, aws.String("RUNNING"))
	if err != nil {
		return nil, err
	}
	stoppedTasks, err := GetAllTasksByDefinitionStatus(cluster, def, aws.String("STOPPED"))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Task Summary: [%d] running, [%d] stopped\n", len(runningTasks), len(stoppedTasks))
	if status == "running" {
		return runningTasks, nil
	}
	if status == "stopped" {
		return stoppedTasks, nil
	}
	return append(runningTasks, stoppedTasks...), nil
}

// GetAllTasksByDefinitionStatus gets all of the tasks of a certain status
// for a task definition
func GetAllTasksByDefinitionStatus(cluster string, def *ecs.TaskDefinition, status *string) ([]*ecs.Task, error) {
	svc := assertECS()
	params := &ecs.ListTasksInput{
		Cluster:       aws.String(cluster),
		Family:        def.Family,
		DesiredStatus: status,
	}
	allArns := make([]*string, 0)
	svc.ListTasksPages(params, func(page *ecs.ListTasksOutput, lastPage bool) bool {
		allArns = append(allArns, page.TaskArns...)
		return !lastPage
	})
	maxTasks := len(allArns)
	if maxTasks > 100 {
		maxTasks = 100
	}
	if maxTasks == 0 {
		return make([]*ecs.Task, 0), nil
	}
	tasks, err := svc.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   allArns[:maxTasks],
	})
	if err != nil {
		return nil, err
	}
	return tasks.Tasks, nil
}

// GetLogs returns cloudwatch logs for a specific set of log streams
// matching a pattern within a log group and region
func GetLogs(cluster, service, status string) error {
	def, err := GetCurrentTaskDefinition(cluster, service)
	if err != nil {
		return err
	}
	allTasks, err := GetAllTasksByDefinition(cluster, def, status)
	if err != nil {
		return fmt.Errorf("Problem getting tasks by definition: %v", err)
	}
	logConfig := def.ContainerDefinitions[0].LogConfiguration
	if *logConfig.LogDriver != "awslogs" {
		return fmt.Errorf("GetLogs function requires awslogs driver, found %v", def.ContainerDefinitions[0].LogConfiguration)
	}
	group := logConfig.Options["awslogs-group"]
	prefix := logConfig.Options["awslogs-stream-prefix"]
	region := *logConfig.Options["awslogs-region"]
	allStreams := make([]*cloudwatchlogs.GetLogEventsInput, len(allTasks))
	for i, task := range allTasks {
		taskID := GetTaskIDFromArn(*task.TaskArn)
		allStreams[i] = &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  group,
			LogStreamName: aws.String(fmt.Sprintf("%s/%s/%s", *prefix, *task.Containers[0].Name, taskID)),
		}
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(allStreams))
	allLogsStream := make(chan string)
	for _, params := range allStreams {
		go func(params *cloudwatchlogs.GetLogEventsInput) {
			defer wg.Done()
			logs, err := GetAllLogs(region, params)
			if err == nil {
				for _, log := range logs {
					allLogsStream <- log
				}
			}
		}(params)
	}
	allLogs := make([]string, 0)
	go func() {
		for curr := range allLogsStream {
			allLogs = append(allLogs, curr)
		}
	}()
	wg.Wait()
	sort.Strings(allLogs)
	fmt.Println(strings.Join(allLogs, "\n"))
	return nil
}

// GetAllLogs retrieves the entire log history
func GetAllLogs(region string, params *cloudwatchlogs.GetLogEventsInput) ([]string, error) {
	svc := assertCloudWatch(region)
	allEvents := make([]string, 0)
	err := svc.GetLogEventsPages(params, func(output *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
		for _, event := range output.Events {
			val := time.Unix(*event.Timestamp/1000, 0)
			message := fmt.Sprintf("[%v] %v", val.Format("2006-01-02 15:04:05 MST"), *event.Message)
			allEvents = append(allEvents, message)
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return allEvents, nil
}

// GetTaskIDFromArn returns the last part of an arn
// for an ecs task
func GetTaskIDFromArn(arn string) string {
	parts := strings.Split(arn, "/")
	return parts[len(parts)-1]
}

// GetTaskLogs prints logs to stdout
func GetTaskLogs(def *ecs.TaskDefinition, taskID string) error {
	logConfig := def.ContainerDefinitions[0].LogConfiguration
	if *logConfig.LogDriver != "awslogs" {
		return fmt.Errorf("logs command requires awslogs driver, found %v", def.ContainerDefinitions[0].LogConfiguration)
	}
	group := logConfig.Options["awslogs-group"]
	prefix := logConfig.Options["awslogs-stream-prefix"]
	region := *logConfig.Options["awslogs-region"]
	svc := assertCloudWatch(region)
	pageNum := 0
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        group,
		LogStreamNamePrefix: prefix,
	}
	if taskID != "" {
		params.SetLogStreamNamePrefix(fmt.Sprintf("%s/%s/%s", *prefix, *def.Family, taskID))
	}
	var eventErr error
	err := svc.DescribeLogStreamsPages(params, func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
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
	startTime := time.Now().Add(-10 * time.Minute)
	eventParams := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(name),
		StartTime:     aws.Int64(startTime.Unix() * 1000),
	}
	eventPage := 0
	eventErr := svc.GetLogEventsPages(eventParams, func(events *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
		eventPage++
		for _, event := range events.Events {
			val := time.Unix(*event.Timestamp/1000, 0)
			fmt.Println(fmt.Sprintf("[%v] %v", val.Format("2006-01-02 15:04:05 MST"), *event.Message))
		}
		return eventPage <= 3
	})
	if eventErr != nil {
		return eventErr
	}
	return nil
}

func LocateTaskDef(cluster, service, source string) (*ecs.TaskDefinition, error) {
	if source == "newest" {
		return GetNewestTaskDefinition(cluster, service)
	}
	return GetCurrentTaskDefinition(cluster, service)
}

// RunTaskWithCommand runs a one-off task with a command
// override.
func RunTaskWithCommand(cluster string, task *ecs.TaskDefinition, command string) (*ecs.RunTaskOutput, error) {
	svc := assertECS()
	commandParts, err := parseCommandOverride(command)
	if err != nil {
		return nil, err
	}
	commandPointers := mapCommandStrPointer(commandParts)
	return svc.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(cluster),
		TaskDefinition: task.TaskDefinitionArn,
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				{
					Name:    task.ContainerDefinitions[0].Name,
					Command: commandPointers,
				},
			},
		},
	})
}

func mapCommandStrPointer(parts []string) []*string {
	commandPointers := make([]*string, len(parts))
	for i := 0; i < len(parts); i++ {
		commandPointers[i] = aws.String(strings.Trim(parts[i], " "))
	}
	return commandPointers
}

// FindRoleByName paginates through roles and returns
// any found that match the name string
func FindRoleByName(name string) (*iam.Role, error) {
	svc := assertIAM()
	params := &iam.ListRolesInput{
		MaxItems: aws.Int64(100),
	}
	var found *iam.Role
	err := svc.ListRolesPages(params, func(rolesOutput *iam.ListRolesOutput, lastPage bool) bool {
		for _, role := range rolesOutput.Roles {
			if *role.RoleName == name {
				found = role
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, fmt.Errorf("unable to find role by name: %v", err)
	}
	if found == nil {
		return nil, fmt.Errorf("unable to find role with name: %v", name)
	}
	return found, nil
}

type containerOverride struct {
	ContainerName string   `json:"name"`
	Command       []string `json:"command"`
}

type ecsCommandOverrideJSON struct {
	ContainerOverrides []containerOverride `json:"containerOverrides"`
}

// CreateScheduledTask creates a scheduled task in an ECS cluster
// with the specified paramters
func CreateScheduledTask(cluster, service, taskSuffix, scheduleExpression, command string) error {
	svc := assertCloudWatchEvents()
	clusters, err := assertClusterMap()
	if err != nil {
		return fmt.Errorf("unable create cluster definitions map: %v", err)
	}
	clusterArn := clusters[cluster]
	ruleName := strings.Join([]string{
		cluster,
		service,
		taskSuffix,
	}, "-")
	taskDefinition, err := GetCurrentTaskDefinition(cluster, service)
	if err != nil {
		return fmt.Errorf("unable to find current task definition: %v", err)
	}
	result, err := svc.ListRules(&cloudwatchevents.ListRulesInput{
		NamePrefix: aws.String(ruleName),
		Limit:      aws.Int64(100),
	})
	if err != nil {
		return fmt.Errorf("unable to list event rules: %v", err)
	}
	if len(result.Rules) == 0 {
		fmt.Printf("Creating Scheduled Task: %v\n", ruleName)
	} else {
		fmt.Printf("Updated Scheduled Task: %v\n", ruleName)
	}
	_, err = svc.PutRule(
		&cloudwatchevents.PutRuleInput{
			Name:               aws.String(ruleName),
			ScheduleExpression: aws.String(scheduleExpression),
			Description: aws.String(fmt.Sprintf(
				"Schedule Expression for %v Service in %v ECS Cluster",
				service,
				cluster,
			)),
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create or update event rule: %v", err)
	}
	return createTaskTarget(clusterArn, service, taskDefinition, ruleName, command)
}

// CreatePostDeploymentTaskInput parameterizes the CreatePostDeploymentTask command
type CreatePostDeploymentTaskInput struct {
	SourceCluster string
	SourceService string
	TargetCluster string
	TargetService string
	Command       string
}

// CreatePostDeploymentTask listens for SERVICE_STEADY_STATE events matching the service
// and cluster, and runs a custom command when a deployment is complete
func CreatePostDeploymentTask(input *CreatePostDeploymentTaskInput) error {
	svc := assertCloudWatchEvents()
	clusters, err := assertClusterMap()
	if err != nil {
		return fmt.Errorf("unable create cluster definitions map: %v", err)
	}
	ruleName := strings.Join([]string{
		input.SourceCluster,
		input.SourceService,
		"stable",
		input.TargetCluster,
		input.TargetService,
	}, "-")
	if len(ruleName) > 64 {
		ruleName = ruleName[0:64]
	}
	taskDefinition, err := GetNewestTaskDefinition(input.TargetCluster, input.TargetService)
	if err != nil {
		return fmt.Errorf("unable to find current task definition: %v", err)
	}
	result, err := svc.ListRules(&cloudwatchevents.ListRulesInput{
		NamePrefix: aws.String(ruleName),
		Limit:      aws.Int64(100),
	})
	if err != nil {
		return fmt.Errorf("unable to list event rules: %v", err)
	}
	if len(result.Rules) == 0 {
		fmt.Printf("Creating Post-Deployment Task: %v\n", ruleName)
	} else {
		fmt.Printf("Updating Post-Deployment Task: %v\n", ruleName)
	}
	eventPattern, err := createPostDeploymentPattern(input.SourceCluster, input.SourceService)
	if err != nil {
		return fmt.Errorf("unable to create event pattern: %v", err)
	}
	_, err = svc.PutRule(
		&cloudwatchevents.PutRuleInput{
			Name:         aws.String(ruleName),
			EventPattern: aws.String(eventPattern),
			Description: aws.String(fmt.Sprintf(
				"Post-Deployment Expression for %v Service in %v ECS Cluster",
				input.SourceService,
				input.SourceCluster,
			)),
		},
	)
	if err != nil {
		return fmt.Errorf("unable to create or update event rule: %v", err)
	}
	return createTaskTarget(clusters[input.TargetCluster], input.TargetService, taskDefinition, ruleName, input.Command)
}

func createPostDeploymentPattern(cluster, service string) (string, error) {
	return createEventPattern(cluster, service, "ECS Service Action", "SERVICE_STEADY_STATE")
}

type eventPattern struct {
	Source     []string `json:"source"`
	DetailType []string `json:"detail-type"`
	Resources  []string `json:"resources"`
	Detail     struct {
		EventName []string `json:"eventName"`
	} `json:"detail"`
}

func createEventPattern(cluster, service, detailType, eventType string) (string, error) {
	ecsService, err := FindService(cluster, service)
	if err != nil {
		return "", fmt.Errorf("unable to find service: %v", err)
	}
	pattern := &eventPattern{
		Source:     []string{"aws.ecs"},
		DetailType: []string{detailType},
		Resources:  []string{*ecsService.ServiceArn},
	}
	pattern.Detail.EventName = []string{eventType}
	eventJSON, err := json.Marshal(pattern)
	if err != nil {
		return "", fmt.Errorf("unable to marshal event pattern: %v", err)
	}
	return string(eventJSON), nil
}

func createTaskTarget(clusterArn string, serviceName string, taskDefinition *ecs.TaskDefinition, ruleName, command string) error {
	svc := assertCloudWatchEvents()
	target := &cloudwatchevents.Target{
		Id:      aws.String("1"),
		Arn:     aws.String(clusterArn),
		RoleArn: taskDefinition.TaskRoleArn,
		EcsParameters: &cloudwatchevents.EcsParameters{
			TaskDefinitionArn: taskDefinition.TaskDefinitionArn,
			TaskCount:         aws.Int64(1),
		},
	}
	if target.RoleArn == nil {
		role, err := FindRoleByName("ecsEventsRole")
		if err != nil {
			return err
		}
		target.RoleArn = role.Arn
	}
	if command != "" {
		commands, err := parseCommandOverride(command)
		if err != nil {
			return fmt.Errorf("command syntax invalid: %v", err)
		}
		overrides := ecsCommandOverrideJSON{
			ContainerOverrides: []containerOverride{
				containerOverride{
					ContainerName: *taskDefinition.ContainerDefinitions[0].Name,
					Command:       commands,
				},
			},
		}
		inputJSON, _ := json.Marshal(overrides)
		target.Input = aws.String(string(inputJSON))
	}
	_, err := svc.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule:    aws.String(ruleName),
		Targets: []*cloudwatchevents.Target{target},
	})
	if err != nil {
		return fmt.Errorf("unable to create or update targets %v", err)
	}
	return nil
}

func GetTask(cluster, arn string) (*ecs.Task, error) {
	svc := assertECS()
	result, err := svc.DescribeTasks(&ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []*string{aws.String(arn)},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Tasks) == 0 {
		return nil, fmt.Errorf("task %s not found", arn)
	}
	return result.Tasks[0], nil
}

// TaskExitCode  will tell you if the exit code of a task
func TaskExitCode(task *ecs.Task) (*int64, error) {
	svc := assertECS()
	input := &ecs.DescribeTasksInput{
		Cluster: task.ClusterArn,
		Tasks:   []*string{task.TaskArn},
	}
	output, err := svc.DescribeTasks(input)
	if err != nil {
		return nil, fmt.Errorf("Unable to poll tasks: %v", err)
	}
	if len(output.Tasks) != 1 {
		return nil, fmt.Errorf(
			"Unable to find task (cluster: %v, arn: %v)",
			task.ClusterArn,
			task.TaskArn,
		)
	}
	if len(output.Failures) > 0 {
		concat := ""
		for _, f := range output.Failures {
			concat += f.String()
		}
		var exit int64 = 1
		return &exit, fmt.Errorf("%s", concat)
	}
	var exitCode int64 = 0
	log.Println("status:", *task.LastStatus)
	if *task.LastStatus == "STOPPED" {
		if *task.StopCode == "TaskFailedToStart" {
			exitCode = 1
			err = fmt.Errorf("task exited with error %s", task.String())
		}
		return &exitCode, err
	}
	outputTask := output.Tasks[0]
	for _, container := range outputTask.Containers {
		if container.ExitCode == nil {
			return nil, nil
		}
		if *container.ExitCode > 0 {
			return container.ExitCode, nil
		}
		exitCode = *container.ExitCode
	}
	return &exitCode, nil
}

// CreateRefreshDeployment forces a new deployment on a service.
func CreateRefreshDeployment(cluster, service string) error {
	svc := assertECS()
	_, err := svc.UpdateService(&ecs.UpdateServiceInput{
		Cluster:            aws.String(cluster),
		Service:            aws.String(service),
		ForceNewDeployment: aws.Bool(true),
	})
	return err
}

func PrintTaskURLs(clusterName string, service *ecs.Service) ([]string, error) {
	svc := assertECS()
	tasks, err := svc.ListTasks(&ecs.ListTasksInput{
		Cluster:     service.ClusterArn,
		ServiceName: service.ServiceName,
		MaxResults:  aws.Int64(100),
	})
	if err != nil {
		return nil, err
	}
	urls := []string{}
	for _, taskARNStr := range tasks.TaskArns {
		taskARN, err := arn.Parse(*taskARNStr)
		if err != nil {
			continue
		}
		parts := strings.Split(taskARN.Resource, "/")
		taskId := parts[len(parts)-1]
		urls = append(urls, fmt.Sprintf(
			"https://%s.console.aws.amazon.com/ecs/v2/clusters/%s/services/%s/configuration/%s/configuration/containers/%s?region=%s",
			taskARN.Region,
			clusterName,
			*service.ServiceName,
			taskId,
			*service.ServiceName,
			taskARN.Region,
		))
	}
	return urls, nil
}
