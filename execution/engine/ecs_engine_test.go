package engine

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapter"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"testing"
)

type mockQueueManager struct {
	statusUpdates []string
}

func (mqm *mockQueueManager) Name() string {
	return "mock"
}

func (mqm *mockQueueManager) QurlFor(name string, prefixed bool) (string, error) {
	return "", nil
}

func (mqm *mockQueueManager) Initialize(config.Config) error {
	return nil
}

func (mqm *mockQueueManager) Enqueue(qURL string, run state.Run) error {
	return nil
}

func (mqm *mockQueueManager) ReceiveRun(qURL string) (queue.RunReceipt, error) {
	return queue.RunReceipt{}, nil
}

func (mqm *mockQueueManager) ReceiveStatus(qURL string) (queue.StatusReceipt, error) {
	popped := mqm.statusUpdates[0]
	mqm.statusUpdates = mqm.statusUpdates[1:]

	return queue.StatusReceipt{StatusUpdate: &popped}, nil
}

func (mqm *mockQueueManager) List() ([]string, error) {
	return nil, nil
}

type mockSQSClient struct {
	queueArn string
}

func (msqs *mockSQSClient) GetQueueAttributes(input *sqs.GetQueueAttributesInput) (*sqs.GetQueueAttributesOutput, error) {
	return &sqs.GetQueueAttributesOutput{
		Attributes: map[string]*string{
			"QueueArn": &msqs.queueArn,
		},
	}, nil
}

type mockCloudWatchClient struct {
}

func (mcwc *mockCloudWatchClient) PutRule(input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
	return &cloudwatchevents.PutRuleOutput{
		RuleArn: aws.String("ruleArn"),
	}, nil
}

func (mcwc *mockCloudWatchClient) PutTargets(input *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {
	return &cloudwatchevents.PutTargetsOutput{
		FailedEntryCount: aws.Int64(0),
	}, nil
}

type testClient struct {
	t               *testing.T
	instanceID      string
	instanceDNSName string
}

func (tc *testClient) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	if input.Cluster == nil || len(*input.Cluster) == 0 {
		tc.t.Errorf("Expected non-nil and non-empty cluster name")
	}

	if len(input.ContainerInstances) != 1 {
		tc.t.Errorf("Expected to describe exactly one container instance")
	}

	ci := ecs.ContainerInstance{
		Ec2InstanceId: &tc.instanceID,
	}
	res := ecs.DescribeContainerInstancesOutput{
		ContainerInstances: []*ecs.ContainerInstance{&ci},
	}
	return &res, nil
}

func (tc *testClient) DescribeInstances(input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	if len(input.InstanceIds) != 1 {
		tc.t.Errorf("Expected exactly one instance id to describe")
	}

	instance := ec2.Instance{
		PrivateDnsName: &tc.instanceDNSName,
	}

	rsv := ec2.Reservation{
		Instances: []*ec2.Instance{&instance},
	}

	return &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{&rsv},
	}, nil
}

func setUp(t *testing.T) ECSExecutionEngine {
	confDir := "../../conf"
	conf, _ := config.NewConfig(&confDir)

	qm := &mockQueueManager{
		statusUpdates: []string{
			`{"version":"0","id":"cupcake","detail-type":"ECS Task State Change","source":"aws.ecs","account":"acctid","time":"2017-12-02T15:06:37Z","region":"us-east-1","resources":["taskarn"],"detail":{"clusterArn":"clusterarn","containerInstanceArn":"ciarn","containers":[{"containerArn":"containerarn","exitCode":0,"lastStatus":"STOPPED","name":"name","taskArn":"arn1"}],"createdAt":"2017-12-02T15:00:06.261Z","desiredStatus":"STOPPED","group":"grp","lastStatus":"STOPPED","overrides":{"containerOverrides":[{"environment":[{"name":"FLOTILLA_SERVER_MODE","value":"prod"},{"name":"FLOTILLA_RUN_ID","value":"runid1"}],"name":"name"}]},"startedAt":"2017-12-02T15:00:08.922Z","startedBy":"starterthing","stoppedAt":"2017-12-02T15:06:37.429Z","stoppedReason":"Essential container in task exited","updatedAt":"2017-12-02T15:06:37.429Z","taskArn":"arn1","taskDefinitionArn":"def1","version":3}}`,
		},
	}

	client := testClient{
		t:               t,
		instanceID:      "cupcake",
		instanceDNSName: "sprinkles",
	}

	a, _ := adapter.NewECSAdapter(conf, &client, &client)

	eng := ECSExecutionEngine{
		qm:        qm,
		sqsClient: &mockSQSClient{"qArn"},
		cwClient:  &mockCloudWatchClient{},
	}
	eng.Initialize(conf)
	eng.adapter = a

	return eng
}

func TestECSExecutionEngine_PollStatus(t *testing.T) {
	eng := setUp(t)

	r, err := eng.PollStatus()
	if err != nil {
		t.Error(err)
	}

	run := r.Run
	if run.Status != state.StatusStopped {
		t.Errorf("Expected status: %s but wa %s", state.StatusStopped, run.Status)
	}

	if run.ExitCode == nil {
		t.Errorf("Expected non nil exit code")
	}

	if run.TaskArn != "arn1" {
		t.Errorf("Expected task arn: [arn1] but was %s", run.TaskArn)
	}
}
