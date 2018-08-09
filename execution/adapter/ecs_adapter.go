package adapter

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

//
// ECSAdapter translates back and forth from ECS api objects to our representation
//
type ECSAdapter interface {
	AdaptTask(task ecs.Task) state.Run
	AdaptRun(definition state.Definition, run state.Run) ecs.RunTaskInput
	AdaptDefinition(definition state.Definition) ecs.RegisterTaskDefinitionInput
	AdaptTaskDef(taskDef ecs.TaskDefinition) state.Definition
}

type EC2ServiceClient interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

type ECSServiceClient interface {
	DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
}

type ecsAdapter struct {
	ecsClient ECSServiceClient
	ec2Client EC2ServiceClient
	conf      config.Config
	retriable []string
}

var mainContainerName = "main"

//
// NewECSAdapter configures and returns an ecs adapter for translating
// from ECS api specific objects to our representation
//
func NewECSAdapter(conf config.Config, ecsClient ECSServiceClient, ec2Client EC2ServiceClient) (ECSAdapter, error) {
	adapter := ecsAdapter{
		conf:      conf,
		ec2Client: ec2Client,
		ecsClient: ecsClient,
		retriable: []string{
			"CannotCreateContainerError",
			"CannotStartContainerError",
			"CannotPullContainerError",
		},
	}
	return &adapter, nil
}

//
// AdaptTask converts from an ecs task to a generic run
//
func (a *ecsAdapter) AdaptTask(task ecs.Task) state.Run {
	run := state.Run{
		TaskArn:    *task.TaskArn,
		StartedAt:  task.StartedAt,
		FinishedAt: task.StoppedAt,
	}

	// Ignore error here
	// TODO - we should log warning
	//
	res, _ := a.ecsClient.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster: task.ClusterArn,
		ContainerInstances: []*string{
			task.ContainerInstanceArn,
		},
	})

	if len(res.ContainerInstances) == 1 {
		cinstance := *res.ContainerInstances[0]
		run.InstanceID = *cinstance.Ec2InstanceId
		r, _ := a.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{cinstance.Ec2InstanceId},
		})
		if len(r.Reservations) == 1 && len(r.Reservations[0].Instances) == 1 {
			run.InstanceDNSName = *r.Reservations[0].Instances[0].PrivateDnsName
		}
	}

	mainContainer := a.extractMainContainer(task.Containers)
	mainContainerOverrides := a.extractMainOverrides(task.Overrides.ContainerOverrides)

	if mainContainerOverrides != nil {
		env := mainContainerOverrides.Environment
		if len(env) > 0 {
			runEnv := make([]state.EnvVar, len(env))
			for i, kv := range env {
				runEnv[i] = state.EnvVar{
					Name:  *kv.Name,
					Value: *kv.Value,
				}
			}
			cast := state.EnvList(runEnv)
			run.Env = &cast
		}
	}

	if mainContainer != nil {
		run.ExitCode = mainContainer.ExitCode
		run.Status = *mainContainer.LastStatus
	}

	if task.DesiredStatus != nil && *task.DesiredStatus == state.StatusStopped {
		run.Status = state.StatusStopped
	}

	if a.needsRetried(run, task) {
		run.Status = state.StatusNeedsRetry
		run.InstanceID = ""
		run.InstanceDNSName = ""
	}

	return run
}

func (a *ecsAdapter) needsRetried(run state.Run, task ecs.Task) bool {
	//
	// This is a -strong- indication of abnormal exit, not internal to the run
	//
	if run.Status == state.StatusStopped && run.ExitCode == nil {
		container := a.extractMainContainer(task.Containers)
		if container != nil && container.Reason != nil {
			for _, retriable := range a.retriable {
				// Container's stopped reason contains a retriable error
				if strings.Contains(*container.Reason, retriable) {
					return true
				}
			}
		}
	}
	return false
}

func (a *ecsAdapter) extractMainContainer(containers []*ecs.Container) *ecs.Container {
	// Handle both the legacy case of 1 container and the multiple containers case
	if len(containers) == 1 {
		return containers[0]
	} else {
		for i := range containers {
			if *containers[i].Name == mainContainerName {
				return containers[i]
			}
		}
	}
	return nil
}

func (a *ecsAdapter) extractMainOverrides(overrides []*ecs.ContainerOverride) *ecs.ContainerOverride {
	if len(overrides) == 1 {
		return overrides[0]
	} else {
		for i := range overrides {
			if *overrides[i].Name == mainContainerName {
				return overrides[i]
			}
		}
	}
	return nil
}

//
// AdaptRun translates the definition and run into the required arguments
// to run an ecs task. There are -several- simplifications to be aware of
//
// 1. There is only ever *1* task launched per run at a time
// 2. Only environment variable overrides are supported (think of these as parameters)
//    Once we start copying the definition information (eg. command, memory, cpu) directly
//    onto the run we will make use of these overrides since it's important to run what
//    we asked for -at the time- of run creation
//
func (a *ecsAdapter) AdaptRun(definition state.Definition, run state.Run) ecs.RunTaskInput {
	n := int64(1)

	overrides := ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{a.envOverrides(definition, run)},
	}

	rti := ecs.RunTaskInput{
		Cluster:        &run.ClusterName,
		Count:          &n,
		StartedBy:      aws.String("flotilla"),
		TaskDefinition: &definition.Arn,
		Overrides:      &overrides,
	}
	return rti
}

func (a *ecsAdapter) envOverrides(definition state.Definition, run state.Run) *ecs.ContainerOverride {
	if run.Env == nil {
		return nil
	}

	pairs := make([]*ecs.KeyValuePair, len(*run.Env))
	for i, ev := range *run.Env {
		name := ev.Name
		value := ev.Value
		pairs[i] = &ecs.KeyValuePair{
			Name:  &name,
			Value: &value,
		}
	}

	res := ecs.ContainerOverride{
		Name:        &mainContainerName,
		Environment: pairs,
	}
	return &res
}

//
// AdaptDefinition translates from definition to the ecs arguments for registering a task
//
// Several simplifications and assumptions are made
// * see `defaultContainerDefinition` for chosen defaults regarding user, privileged mode, networking, etc
// * we wrap the command specified to ensure lines are echoed and the exit code is captured and is an injection
//   point for other infra related concerns
//
// TODO - add CPU
//
func (a *ecsAdapter) AdaptDefinition(definition state.Definition) ecs.RegisterTaskDefinitionInput {
	// Get additional containers and corresponding networking links
	additionalContainers := a.additionalContainers()
	links := a.getLinks(additionalContainers)

	// Get main container
	containerDef := a.defaultContainerDefinition()
	containerDef.Links = links // Link additional containers into main container
	containerDef.Image = &definition.Image
	containerDef.Memory = definition.Memory
	containerDef.Name = &mainContainerName
	containerDef.DockerLabels = map[string]*string{
		"alias":      &definition.Alias,
		"group.name": &definition.GroupName,
	}

	cmdString, err := definition.WrappedCommand()
	if err != nil {
		// Fallback
		cmdString = definition.Command
	}
	cmds := []string{"bash", "-l", "-c", cmdString}
	containerDef.Command = []*string{
		&cmds[0], &cmds[1], &cmds[2], &cmds[3],
	}

	if definition.Ports != nil {
		protocol := "tcp"
		containerDef.PortMappings = make([]*ecs.PortMapping, len(*definition.Ports))
		for i, p := range *definition.Ports {
			port := int64(p)
			containerDef.PortMappings[i] = &ecs.PortMapping{
				Protocol:      &protocol,
				HostPort:      &port,
				ContainerPort: &port,
			}
		}
	}

	if definition.Env != nil {
		containerDef.Environment = make([]*ecs.KeyValuePair, len(*definition.Env))
		for i, e := range *definition.Env {
			name := e.Name
			value := e.Value
			containerDef.Environment[i] = &ecs.KeyValuePair{
				Name:  &name,
				Value: &value,
			}
		}
	}

	if definition.Tags != nil {
		tagsList := strings.Join(*definition.Tags, ",")
		containerDef.DockerLabels["tags"] = &tagsList
	}

	networkMode := "bridge"
	containerDefns := []*ecs.ContainerDefinition{containerDef}
	containerDefns = append(containerDefns, additionalContainers...)

	return ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: containerDefns,
		Family:               &definition.DefinitionID,
		NetworkMode:          &networkMode,
	}
}

//
// AdaptTaskDef translates from an ecs task definition to our definition object
// * see `AdaptTaskDefinition` for translation details
//
func (a *ecsAdapter) AdaptTaskDef(taskDef ecs.TaskDefinition) state.Definition {
	adapted := state.Definition{
		Arn:           *taskDef.TaskDefinitionArn,
		DefinitionID:  *taskDef.Family,   // Family==DefinitionID
		ContainerName: mainContainerName, // ContainerName is always `mainContainerName`
	}

	container := a.extractMainDefinition(taskDef.ContainerDefinitions)
	if container != nil {
		adapted.Memory = container.Memory
		adapted.Image = *container.Image

		alias, _ := container.DockerLabels["alias"]
		groupName, _ := container.DockerLabels["group.name"]
		tagList, ok := container.DockerLabels["tags"]
		if ok {
			tagSplits := strings.Split(*tagList, ",")
			tags := state.Tags(tagSplits)
			adapted.Tags = &tags
		}

		adapted.GroupName = *groupName
		adapted.Alias = *alias

		if len(container.PortMappings) > 0 {
			ports := make([]int, len(container.PortMappings))
			for i, pm := range container.PortMappings {
				ports[i] = int(*pm.ContainerPort)
			}
			adaptedPorts := state.PortsList(ports)
			adapted.Ports = &adaptedPorts
		}

		env := make([]state.EnvVar, len(container.Environment))
		for i, e := range container.Environment {
			env[i] = state.EnvVar{
				Name:  *e.Name,
				Value: *e.Value,
			}
		}
		adaptedEnv := state.EnvList(env)
		adapted.Env = &adaptedEnv
	}

	return adapted
}

func (a *ecsAdapter) extractMainDefinition(definitions []*ecs.ContainerDefinition) *ecs.ContainerDefinition {
	if len(definitions) == 1 {
		return definitions[0]
	} else {
		for i := range definitions {
			if *definitions[i].Name == mainContainerName {
				return definitions[i]
			}
		}
	}
	return nil
}

func (a *ecsAdapter) defaultContainerDefinition() *ecs.ContainerDefinition {
	// Default container -must- be essential
	return &ecs.ContainerDefinition{
		Essential:         aws.Bool(true),
		User:              aws.String("root"),
		DisableNetworking: aws.Bool(false),
		Privileged:        aws.Bool(true),
		LogConfiguration:  a.logConfiguration(),
	}
}

func (a *ecsAdapter) additionalContainers() []*ecs.ContainerDefinition {
	res := []*ecs.ContainerDefinition{}
	if !a.conf.IsSet("additional_containers") {
		return res
	}

	fromConf := a.conf.GetStringMapInterfaceSlice("additional_containers")
	for _, c := range fromConf {
		var (
			name    string
			image   string
			user    string
			memory  int64
			cpu     int64
			command string
		)

		name = c["name"].(string)
		image = c["image"].(string)
		if memRaw, ok := c["memory"]; ok {
			memory = int64(memRaw.(int))
		} else {
			memory = 100
		}

		if cpuRaw, ok := c["cpu"]; ok {
			cpu = int64(cpuRaw.(int))
		} else {
			cpu = 100
		}

		if userRaw, ok := c["user"]; ok {
			user = userRaw.(string)
		} else {
			user = "root"
		}

		if cmdRaw, ok := c["command"]; ok {
			command = cmdRaw.(string)
		}

		res = append(res, &ecs.ContainerDefinition{
			Name:              &name,
			Image:             &image,
			Command:           []*string{&command},
			Memory:            &memory,
			Cpu:               &cpu,
			User:              &user,
			Essential:         aws.Bool(false),
			DisableNetworking: aws.Bool(false),
			Privileged:        aws.Bool(true),
			LogConfiguration:  a.logConfiguration(),
		})
	}
	return res
}

func (a *ecsAdapter) logConfiguration() *ecs.LogConfiguration {
	logDriver := a.conf.GetString("log.driver.name")
	if len(logDriver) == 0 {
		logDriver = "awslogs"
	}
	confLogOptions := a.conf.GetStringMapString("log.driver.options")
	logOptions := make(map[string]*string, len(confLogOptions))
	for k, v := range confLogOptions {
		val := v
		logOptions[k] = &val
	}

	//
	// Allow defining log group as -either- log namespace or
	// awslogs-group
	//
	_, ok := logOptions["awslogs-group"]
	if !ok {
		logGroup := a.conf.GetString("log.namespace")
		logOptions["awslogs-group"] = &logGroup
	}

	return &ecs.LogConfiguration{
		LogDriver: &logDriver,
		Options:   logOptions,
	}
}

func (a *ecsAdapter) getLinks(additionalContainers []*ecs.ContainerDefinition) []*string {
	res := make([]*string, len(additionalContainers))
	for i, ac := range additionalContainers {
		name := fmt.Sprintf("%s:%s", *ac.Name, *ac.Name)
		res[i] = &name
	}
	return res
}
