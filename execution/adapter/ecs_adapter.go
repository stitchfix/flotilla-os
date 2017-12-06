package adapter

import (
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
		GroupName:  *task.Group,
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

	if task.Overrides != nil && len(task.Overrides.ContainerOverrides) == 1 {
		env := task.Overrides.ContainerOverrides[0].Environment
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

	if task.DesiredStatus != nil && *task.DesiredStatus == state.StatusStopped {
		run.Status = state.StatusStopped
	} else {
		run.Status = *task.LastStatus
	}

	if len(task.Containers) == 1 {
		run.ExitCode = task.Containers[0].ExitCode
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
		containerReason := "?"
		if len(task.Containers) == 1 {
			container := task.Containers[0]
			if container != nil && container.Reason != nil {
				containerReason = *container.Reason
			}
		}

		for _, retriable := range a.retriable {
			// Container's stopped reason contains a retriable error
			if strings.Contains(containerReason, retriable) {
				return true
			}
		}
	}
	return false
}

//
// AdaptRun translates the definition and run into the required arguments
// to run an ecs task. There are -several- simplifications to be aware of
//
// 1. There is currently only ever *1* container per definition
// 2. There is only ever *1* task launched per run at a time
// 3. Only environment variable overrides are supported (think of these as parameters)
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
		StartedBy:      &definition.GroupName,
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

	//
	// Support legacy case of differing container name and definition id
	//
	containerName := definition.DefinitionID
	if definition.ContainerName != definition.DefinitionID {
		containerName = definition.ContainerName
	}

	res := ecs.ContainerOverride{
		Name:        &containerName,
		Environment: pairs,
	}
	return &res
}

//
// AdaptDefinition translates from definition to the ecs arguments for registering a task
//
// Several simplifications and assumptions are made
// * see `defaultContainerDefinition` for chosen defaults regarding user, privileged mode, networking, etc
// * for now, exactly -one- container per definition is defined AND the DefinitionID == ecs family == container name
// * we -always- use host networking; this dramatically simplifies the reasoning the end-users have to do
//   about the way in which their runs are going to function; esp wrt external libraries and frameworks
//   *** port mappings are maintained as a mechanism to pre-allocate what ports to use and allows some flexibility in
//       networking; MOST runs will not use this currently as we're using "host" networking mode
// * we wrap the command specified to ensure lines are echoed and the exit code is captured and is an injection
//   point for other infra related concerns
//
// TODO - add CPU
//
func (a *ecsAdapter) AdaptDefinition(definition state.Definition) ecs.RegisterTaskDefinitionInput {
	containerDef := a.defaultContainerDefinition()
	containerDef.Image = &definition.Image
	containerDef.Memory = definition.Memory
	containerDef.Name = &definition.DefinitionID
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

	networkMode := "host"
	return ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{containerDef},
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
		DefinitionID:  *taskDef.Family, // Family==ContainerName==DefinitionID
		ContainerName: *taskDef.Family,
	}

	if len(taskDef.ContainerDefinitions) == 1 {
		container := taskDef.ContainerDefinitions[0]

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

func (a *ecsAdapter) defaultContainerDefinition() *ecs.ContainerDefinition {
	essential := true
	user := "root"
	disableNetworking := false
	privileged := true

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

	logConfiguration := ecs.LogConfiguration{
		LogDriver: &logDriver,
		Options:   logOptions,
	}

	return &ecs.ContainerDefinition{
		Essential:         &essential,
		User:              &user,
		DisableNetworking: &disableNetworking,
		Privileged:        &privileged,
		LogConfiguration:  &logConfiguration,
	}
}
