package engine

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sqs"
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
	cwClient   cloudwatchServiceClient
	sqsClient  sqsClient
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

type cloudwatchServiceClient interface {
	PutRule(input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error)
	PutTargets(input *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error)
	ListRuleNamesByTarget(input *cloudwatchevents.ListRuleNamesByTargetInput) (*cloudwatchevents.ListRuleNamesByTargetOutput, error)
}

type sqsClient interface {
	GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error)
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

	if !conf.IsSet("queue.status") {
		return fmt.Errorf("ECSExecutionEngine needs [queue.status] set in config")
	}

	if !conf.IsSet("queue.status_rule") {
		return fmt.Errorf("ECSExecutionEngine needs [queue.status_rule] set in config")
	}

	var (
		adpt adapter.ECSAdapter
		err  error
	)

	flotillaMode := conf.GetString("flotilla_mode")

	//
	// When mode is not test, setup and initialize all aws clients
	// - this isn't ideal; is there another way?
	//
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(conf.GetString("aws_default_region"))}))

		ecsClient := ecs.New(sess)
		ec2Client := ec2.New(sess)

		ee.ecsClient = ecsClient
		ee.cwClient = cloudwatchevents.New(sess)
		ee.sqsClient = sqs.New(sess)
		adpt, err = adapter.NewECSAdapter(conf, ecsClient, ec2Client)
		if err != nil {
			return err
		}
	}

	ee.adapter = adpt

	if ee.qm == nil {
		return fmt.Errorf("No queue.Manager implementation; ECSExecutionEngine needs a queue.Manager")
	}

	//
	// Calling QurlFor creates the status queue if it does not exist
	// - this is necessary for the next step of creating an ecs
	//   task update rule in cloudwatch which routes task updates
	//   to the status queue
	//
	statusQueue := conf.GetString("queue.status")
	ee.statusQurl, err = ee.qm.QurlFor(statusQueue, false)
	if err != nil {
		return err
	}

	statusRule := conf.GetString("queue.status_rule")
	return ee.createOrUpdateEventRule(statusRule, statusQueue)
}

func (ee *ECSExecutionEngine) createOrUpdateEventRule(statusRule string, statusQueue string) error {
	_, err := ee.cwClient.PutRule(&cloudwatchevents.PutRuleInput{
		Description:  aws.String("Routes ecs task status events to flotilla status queues"),
		Name:         &statusRule,
		EventPattern: aws.String(`{"source":["aws.ecs"],"detail-type":["ECS Task State Change"]}`),
	})

	if err != nil {
		return err
	}

	// Route status events to the status queue
	targetArn, err := ee.getTargetArn(ee.statusQurl)
	if err != nil {
		return fmt.Errorf("Error getting target arn for %s; message: [%s]", ee.statusQurl, err.Error())
	}

	names, err := ee.cwClient.ListRuleNamesByTarget(&cloudwatchevents.ListRuleNamesByTargetInput{
		TargetArn: &targetArn,
	})
	if err != nil {
		return fmt.Errorf("Error listing rules for target: [%s]; message: [%s]", targetArn, err.Error())
	}

	if len(names.RuleNames) > 0 && *names.RuleNames[0] == statusRule {
		return nil
	}

	res, err := ee.cwClient.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: &statusRule,
		Targets: []*cloudwatchevents.Target{
			{
				Arn: &targetArn,
				Id:  &statusQueue,
			},
		},
	})

	if err != nil {
		return err
	}

	if *res.FailedEntryCount > 0 {
		failed := res.FailedEntries[0]
		return fmt.Errorf("Error creating routing rule for ecs status messages [%s]", *failed.ErrorMessage)
	}

	return nil
}

func (ee *ECSExecutionEngine) getTargetArn(qurl string) (string, error) {
	var arn string
	res, err := ee.sqsClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: &qurl,
		AttributeNames: []*string{
			aws.String("QueueArn"),
		},
	})
	if err != nil {
		return arn, err
	}
	if res.Attributes["QueueArn"] != nil {
		return *res.Attributes["QueueArn"], nil
	}
	return arn, fmt.Errorf("Couldn't get queue arn")
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

	//
	// If we receive an update that is empty, don't try to deserialize it
	//
	if rawReceipt.StatusUpdate != nil {
		err = json.Unmarshal([]byte(*rawReceipt.StatusUpdate), &update)
		if err != nil {
			return receipt, fmt.Errorf("Error: %v\nJSON: [%s]", err.Error(), rawReceipt.StatusUpdate)
		}
		adapted := ee.adapter.AdaptTask(update.Detail)
		receipt.Run = &adapted
	}

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
			continue
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
