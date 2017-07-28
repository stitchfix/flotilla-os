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

type ECSAdapter interface {
	AdaptTask(task ecs.Task) state.Run
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

	if task.DesiredStatus != nil && *task.DesiredStatus == "STOPPED" {
		run.Status = "STOPPED"
	} else {
		run.Status = *task.LastStatus
	}

	if len(task.Containers) == 1 {
		run.ExitCode = task.Containers[0].ExitCode
	}

	if a.needsRetried(run, task) {
		run.Status = "NEEDS_RETRY"
		run.InstanceID = ""
		run.InstanceDNSName = ""
	}

	return run
}

func (a *ecsAdapter) needsRetried(run state.Run, task ecs.Task) bool {
	//
	// This is a -strong- indication of abnormal exit, not internal to the run
	//
	if run.Status == "STOPPED" && run.ExitCode == nil {
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
