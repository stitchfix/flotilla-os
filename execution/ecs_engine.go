package execution

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
// ECSExecutionEngine submits runs to ecs
//
type ECSExecutionEngine struct {
	ecsClient ecsServiceClient
	ec2Client ec2ServiceClient
}

type ecsServiceClient interface {
	RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error)
	DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
}

type ec2ServiceClient interface {
	DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
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
		ee.ec2Client = ec2.New(sess)
	}
	return nil
}

func (ee *ECSExecutionEngine) Execute(definition state.Definition, run state.Run) (state.Run, error) {
	var executed state.Run
	result, err := ee.ecsClient.RunTask(&ee.toRunTaskInput(definition, run))
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
	run := state.Run{
		TaskArn:    *task.TaskArn,
		GroupName:  *task.Group,
		StartedAt:  task.StartedAt,
		FinishedAt: task.StoppedAt,
	}

	// Ignore error here
	// TODO - we should log warning
	//
	res, _ := ee.ecsClient.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster: task.ClusterArn,
		ContainerInstances: []*string{
			task.ContainerInstanceArn,
		},
	})

	if len(res.ContainerInstances) == 1 {
		cinstance := *res.ContainerInstances[0]
		run.InstanceID = *cinstance.Ec2InstanceId
		r, _ := ee.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{cinstance.Ec2InstanceId},
		})
		if len(r.Reservations) == 1 && len(r.Reservations[0].Instances) == 1 {
			run.InstanceDNSName = *r.Reservations[0].Instances[0].PrivateDnsName
		}
	}

	if task.Overrides != nil && len(task.Overrides.ContainerOverrides) > 0 {
		// map to env vars
	}

	if *task.DesiredStatus == "STOPPED" {
		run.Status = "STOPPED"
	} else {
		run.Status = *task.LastStatus
	}

	if len(task.Containers) == 1 {
		run.ExitCode = task.Containers[0].ExitCode
	}

	if ee.needsRetried(run, task) {

	}

	return run
}

func (ee *ECSExecutionEngine) needsRetried(run state.Run, task ecs.Task) bool {
	/*retry = False
		if cluster_task.status == 'STOPPED' and cluster_task.exit_code is None:
		task_reason = run_info.get('stoppedReason')
		container_reason = '?'
		if 'containers' in run_info and len(run_info.get('containers')) == 1:
		container = run_info.get('containers')[0]
		container_reason = container.get('reason', '?')

		logger.warn(
			"Got STOPPED task: [{t}] with empty exit code, container reason: [{r}], task reason: [{tr}]".format(
				t=cluster_task.run_id, r=container_reason, tr=task_reason))

		codes = [
			'CannotCreateContainerError',
		'CannotStartContainerError',
		'CannotPullContainerError'
	            ]
		retry = True if any([c in container_reason for c in codes]) else False*/
	return false
}
