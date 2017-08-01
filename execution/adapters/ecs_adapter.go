package adapters

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
}

type ec2ServiceClient interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

type ecsServiceClient interface {
	DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
}

type ecsAdapter struct {
	ecsClient ecsServiceClient
	ec2Client ec2ServiceClient
	retriable []string
}

//
// NewECSAdapter configures and returns an ecs adapter for translating
// from ECS api specific objects to our representation
//
func NewECSAdapter(conf config.Config) (ECSAdapter, error) {
	adapter := ecsAdapter{}

	if !conf.IsSet("aws_default_region") {
		return &adapter, fmt.Errorf("ECSAdapter needs [aws_default_region] set in config")
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(conf.GetString("aws_default_region"))}))

		adapter.ecsClient = ecs.New(sess)
		adapter.ec2Client = ec2.New(sess)
	}

	adapter.retriable = []string{
		"CannotCreateContainerError",
		"CannotStartContainerError",
		"CannotPullContainerError",
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
			containerReason = *task.Containers[0].Reason
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
		StartedBy:      &run.GroupName,
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
		pairs[i] = &ecs.KeyValuePair{
			Name:  &ev.Name,
			Value: &ev.Value,
		}
	}

	res := ecs.ContainerOverride{
		Name:        &definition.ContainerName,
		Environment: pairs,
	}
	return &res
}
