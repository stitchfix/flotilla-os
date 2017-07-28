package engine

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapters"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

//
// ECSExecutionEngine submits runs to ecs
//
type ECSExecutionEngine struct {
	ecsClient ecsServiceClient
	adapter   adapters.ECSAdapter
}

type ecsServiceClient interface {
	RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error)
	DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
}

func (ee *ECSExecutionEngine) Initialize(conf config.Config) error {
	if !conf.IsSet("aws_default_region") {
		return fmt.Errorf("ECSExecutionEngine needs [aws_default_region] set in config")
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(conf.GetString("aws_default_region"))}))

		ee.ecsClient = ecs.New(sess)
	}

	adapter, err := adapters.NewECSAdapter(conf)
	if err != nil {
		return err
	}
	ee.adapter = adapter
	return nil
}

func (ee *ECSExecutionEngine) Execute(definition state.Definition, run state.Run) (state.Run, error) {
	var executed state.Run
	rti := ee.toRunTaskInput(definition, run)
	result, err := ee.ecsClient.RunTask(&rti)
	if err != nil {
		return executed, err
	}

	if len(result.Failures) != 0 {
		msg := make([]string, len(result.Failures))
		for i, failure := range result.Failures {
			msg[i] = *failure.Reason
		}
		return executed, fmt.Errorf("ERRORS: %s", strings.Join(msg, "\n"))
	}

	return ee.translateTask(*result.Tasks[0]), nil
}

//
// toRunTaskInput translates the definition and run into the required arguments
// to run an ecs task. There are -several- simplifications to be aware of
//
// 1. There is currently only ever *1* container per definition
// 2. There is only ever *1* task launched per run at a time
// 3. Only environment variable overrides are supported (think of these as parameters)
//    Once we start copying the definition information (eg. command, memory, cpu) directly
//    onto the run we will make use of these overrides since it's important to run what
//    we asked for -at the time- of run creation
//
func (ee *ECSExecutionEngine) toRunTaskInput(definition state.Definition, run state.Run) ecs.RunTaskInput {
	n := int64(1)

	overrides := ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{ee.envOverrides(definition, run)},
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

func (ee *ECSExecutionEngine) envOverrides(definition state.Definition, run state.Run) *ecs.ContainerOverride {
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

func (ee *ECSExecutionEngine) translateTask(task ecs.Task) state.Run {
	return ee.adapter.AdaptTask(task)
}
