package engine

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapter"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

//
// ECSExecutionEngine submits runs to ecs
//
type ECSExecutionEngine struct {
	ecsClient  ecsServiceClient
	adapter    adapter.ECSAdapter
	qm         queue.Manager
	statusQurl string
}

type ecsServiceClient interface {
	RunTask(input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error)
	StopTask(input *ecs.StopTaskInput) (*ecs.StopTaskOutput, error)
	DeregisterTaskDefinition(input *ecs.DeregisterTaskDefinitionInput) (*ecs.DeregisterTaskDefinitionOutput, error)
	RegisterTaskDefinition(input *ecs.RegisterTaskDefinitionInput) (*ecs.RegisterTaskDefinitionOutput, error)
	DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error)
}

type ecsUpdate struct {
	Detail ecs.Task `json:"detail"`
}

//
// Initialize configures the ECSExecutionEngine and initializes internal clients
//
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

	adapter, err := adapter.NewECSAdapter(conf)
	if err != nil {
		return err
	}
	ee.adapter = adapter

	//
	// Get queue manager for queuing runs
	//
	qm, err := queue.NewQueueManager(conf)
	if err != nil {
		return err
	}
	ee.qm = qm

	statusQueue := conf.GetString("queue.status")
	ee.statusQurl, err = ee.qm.QurlFor(statusQueue, false)
	if err != nil {
		return err
	}

	return nil
}

//
// PollStatus pops status updates from the status queue using the QueueManager
//
func (ee *ECSExecutionEngine) PollStatus() (RunReceipt, error) {
	var (
		receipt RunReceipt
		update  ecsUpdate
		err     error
	)

	rawReceipt, err := ee.qm.ReceiveStatus(ee.statusQurl)
	if err != nil {
		return receipt, err
	}

	err = json.Unmarshal([]byte(rawReceipt.StatusUpdate), &update)
	if err != nil {
		return receipt, err
	}

	adapted := ee.adapter.AdaptTask(update.Detail)

	receipt.Run = &adapted
	receipt.Done = rawReceipt.Done
	return receipt, nil
}

//
// PollRuns receives -at most- one run per queue that is pending execution
//
func (ee *ECSExecutionEngine) PollRuns() ([]RunReceipt, error) {
	queues, err := ee.qm.List()
	if err != nil {
		return nil, err
	}

	var runs []RunReceipt
	for _, qurl := range queues {
		//
		// Get new queued Run
		//
		runReceipt, err := ee.qm.ReceiveRun(qurl)
		if err != nil {
			return runs, err
		}

		if runReceipt.Run == nil {
			return runs, err
		}

		runs = append(runs, RunReceipt{runReceipt})
	}
	return runs, nil
}

//
// Enqueue pushes a run onto the queue using the QueueManager
//
func (ee *ECSExecutionEngine) Enqueue(run state.Run) error {
	// Get qurl
	qurl, err := ee.qm.QurlFor(run.ClusterName, true)
	if err != nil {
		return err
	}

	// Queue run
	return ee.qm.Enqueue(qurl, run)
}

//
// Execute takes a pre-configured run and definition and submits them for execution
// to AWS ECS
//
func (ee *ECSExecutionEngine) Execute(definition state.Definition, run state.Run) (state.Run, bool, error) {
	var executed state.Run
	rti := ee.toRunTaskInput(definition, run)
	result, err := ee.ecsClient.RunTask(&rti)
	if err != nil {
		retryable := false
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == ecs.ErrCodeInvalidParameterException {
				if strings.Contains(aerr.Message(), "no container instances") {
					retryable = true
				}
			}
		}
		return executed, retryable, err
	}
	if len(result.Failures) != 0 {
		msg := make([]string, len(result.Failures))
		for i, failure := range result.Failures {
			msg[i] = *failure.Reason
		}
		//
		// Retry these, they are very rare;
		// our upfront validation catches the obvious image and cluster resources
		//
		// IMPORTANT - log these messages
		//
		return executed, true, fmt.Errorf("ERRORS: %s", strings.Join(msg, "\n"))
	}

	return ee.translateTask(*result.Tasks[0]), false, nil
}

//
// Terminate takes a valid run and stops it
//
func (ee *ECSExecutionEngine) Terminate(run state.Run) error {
	_, err := ee.ecsClient.StopTask(&ecs.StopTaskInput{
		Cluster: &run.ClusterName,
		Task:    &run.TaskArn,
	})
	return err
}

//
// Define creates or updates a task definition with ecs
//
func (ee *ECSExecutionEngine) Define(definition state.Definition) (state.Definition, error) {
	rti := ee.adapter.AdaptDefinition(definition)
	result, err := ee.ecsClient.RegisterTaskDefinition(&rti)
	if err != nil {
		return state.Definition{}, err
	}

	//
	// We wrap the command of a definition before registering it with
	// ECS. What this means is that the command returned from registration
	// contains only the *wrapped* version. Reversing the wrapping process
	// using string parsing is brittle. Instead, we make the following
	// assumptions:
	//
	// * Definitions are pre-validated using their `IsValid` method meaning
	//   they must have a non-empty user command
	// * Registering a task definition with ECS does not mutate the user command
	// ** The command acknowledged by ECS is -exactly- the wrapped version
	//    of the command contained in the passed in Definition
	// Hence it should be safe to simply attach the passed in definition's
	// Command field to the output.
	//
	defined := ee.adapter.AdaptTaskDef(*result.TaskDefinition)
	defined.Command = definition.Command
	return defined, nil
}

//
// Deregister deregisters the task definition from ecs
//
func (ee *ECSExecutionEngine) Deregister(definition state.Definition) error {
	_, err := ee.ecsClient.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: &definition.Arn,
	})
	return err
}

func (ee *ECSExecutionEngine) toRunTaskInput(definition state.Definition, run state.Run) ecs.RunTaskInput {
	return ee.adapter.AdaptRun(definition, run)
}

func (ee *ECSExecutionEngine) translateTask(task ecs.Task) state.Run {
	return ee.adapter.AdaptTask(task)
}
