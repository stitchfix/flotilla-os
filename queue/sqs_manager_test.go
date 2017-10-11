package queue

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"testing"
)

type testSQSClient struct {
	t      *testing.T
	queues []*string
	calls  []string
}

func (qc *testSQSClient) GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {
	qc.calls = append(qc.calls, "GetQueueUrl")
	if input.QueueName == nil || len(*input.QueueName) == 0 {
		qc.t.Errorf("Expected non-nil and non empty QueueName")
	}

	if *input.QueueName == "qtest-nope" {
		return nil, errors.New("No queue here")
	}

	qurl := "cupcake"
	return &sqs.GetQueueUrlOutput{QueueUrl: &qurl}, nil
}

func (qc *testSQSClient) CreateQueue(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error) {
	qc.calls = append(qc.calls, "CreateQueue")
	if input.QueueName == nil || len(*input.QueueName) == 0 {
		qc.t.Errorf("Expected non-nil and non empty QueueName")
	}

	if _, ok := input.Attributes["MessageRetentionPeriod"]; !ok {
		qc.t.Errorf("Expected MessageRetentionPeriod in attributes")
	}

	if _, ok := input.Attributes["VisibilityTimeout"]; !ok {
		qc.t.Errorf("Expected VisibilityTimeout in attributes")
	}

	qurl := "nope"
	return &sqs.CreateQueueOutput{QueueUrl: &qurl}, nil
}

func (qc *testSQSClient) ListQueues(input *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
	qc.calls = append(qc.calls, "ListQueues")
	if input.QueueNamePrefix == nil {
		qc.t.Errorf("Expected non-nil QueueNamePrefix")
	}

	if len(*input.QueueNamePrefix) == 0 {
		qc.t.Errorf("Expected non-empty QueueNamePrefix")
	}

	response := sqs.ListQueuesOutput{QueueUrls: qc.queues}
	return &response, nil
}

func (qc *testSQSClient) SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	qc.calls = append(qc.calls, "SendMessage")
	if input.QueueUrl == nil {
		qc.t.Errorf("Expected non-nil QueueUrl")
	}

	if len(*input.QueueUrl) == 0 {
		qc.t.Errorf("Expected non-empty QueueUrl")
	}

	body := input.MessageBody
	if body == nil {
		qc.t.Errorf("Expected non-nil MessageBody")
	}
	var run state.Run
	var smo sqs.SendMessageOutput
	err := json.Unmarshal([]byte(*body), &run)
	if err != nil {
		qc.t.Errorf("Error deserializing MessageBody to Run, [%v]", err)
	}

	if len(run.RunID) == 0 {
		qc.t.Errorf("RunID of deserialized Run should not be empty")
	}
	return &smo, nil
}

func (qc *testSQSClient) ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	qc.calls = append(qc.calls, "ReceiveMessage")
	if input.VisibilityTimeout == nil {
		qc.t.Errorf("Expected non-nil VisibilityTimeout")
	}
	if input.MaxNumberOfMessages == nil {
		qc.t.Errorf("Expected non-nil MaxNumberOfMessages")
	}
	if *input.MaxNumberOfMessages != 1 {
		qc.t.Errorf("Expected MaxNumberOfMessages to be 1, was %v", *input.MaxNumberOfMessages)
	}
	if input.QueueUrl == nil {
		qc.t.Errorf("Expected non-nil QueueUrl")
	}
	if len(*input.QueueUrl) == 0 {
		qc.t.Errorf("Expected non-empty QueueUrl")
	}

	handle := "handle"
	asString := ""
	if *input.QueueUrl == "statusQ" {
		asString = `{"detail":{"taskArn":"sometaskarn","lastStatus":"STOPPED","version":17, "overrides":{"containerOverrides":[{"environment":[{"name":"FLOTILLA_SERVER_MODE","value":"prod"}]}]}}}`
	} else {
		jsonRun, _ := json.Marshal(state.Run{RunID: "cupcake"})
		asString = string(jsonRun)
	}

	msg := sqs.Message{
		ReceiptHandle: &handle,
		Body:          &asString,
	}
	rmo := sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{&msg},
	}
	return &rmo, nil
}

func (qc *testSQSClient) DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {
	qc.calls = append(qc.calls, "DeleteMessage")
	if input.QueueUrl == nil {
		qc.t.Errorf("Expected non-nil QueueUrl")
	}
	if len(*input.QueueUrl) == 0 {
		qc.t.Errorf("Expected non-empty QueueUrl")
	}
	if input.ReceiptHandle == nil {
		qc.t.Errorf("Expected non-nil ReceiptHandle")
	}
	if len(*input.ReceiptHandle) == 0 {
		qc.t.Errorf("Expected non-empty ReceiptHandle")
	}
	return &sqs.DeleteMessageOutput{}, nil
}

func setUp(t *testing.T) SQSManager {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)

	qm := SQSManager{}
	qm.Initialize(c)
	qm.namespace = "qtest"

	qA := "A"
	qB := "B"
	qC := "C"
	qStatus := "statusQ"
	testClient := testSQSClient{
		t:      t,
		queues: []*string{&qA, &qB, &qC, &qStatus},
	}
	qm.qc = &testClient

	return qm
}

func TestSQSManager_List(t *testing.T) {
	qm := setUp(t)

	listed, _ := qm.List()
	if len(listed) != 4 {
		t.Errorf("Expected listed queues to be [4] but was %v", len(listed))
	}
}

func TestSQSManager_Enqueue(t *testing.T) {
	qm := setUp(t)

	var err error
	toQ := state.Run{
		RunID: "cupcake",
	}
	qm.Enqueue("A", toQ)

	err = qm.Enqueue("", toQ)
	if err == nil {
		t.Errorf("Expected empty queue url to result in error")
	}
}

func TestSQSManager_QurlFor(t *testing.T) {
	qm := setUp(t)

	testClient := testSQSClient{t: t}
	qm.qc = &testClient

	expectedCalls := map[string]bool{
		"GetQueueUrl": true,
	}
	qm.QurlFor("cupcake", true)

	if len(testClient.calls) != len(expectedCalls) {
		t.Errorf(
			"Expected exactly %v calls for existing queue, but was %v",
			len(expectedCalls), len(testClient.calls))
	}

	for _, call := range testClient.calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call for existing queue [%v]", call)
		}
	}

	testClient = testSQSClient{t: t}
	qm.qc = &testClient

	expectedCalls = map[string]bool{
		"GetQueueUrl": true,
		"CreateQueue": true,
	}
	qm.QurlFor("nope", true)

	if len(testClient.calls) != len(expectedCalls) {
		t.Errorf(
			"Expected exactly %v calls for non-existing queue, but was %v",
			len(expectedCalls), len(testClient.calls))
	}

	for _, call := range testClient.calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call for non-existing queue [%v]", call)
		}
	}
}

func TestSQSManager_ReceiveRun(t *testing.T) {
	qm := setUp(t)
	receipt, _ := qm.ReceiveRun("A")
	receipt.Done()
}

func TestSQSManager_ReceiveStatus(t *testing.T) {
	qm := setUp(t)
	receipt, _ := qm.ReceiveStatus("statusQ")

	srvMode, ok := receipt.StatusUpdate.GetEnvVar("FLOTILLA_SERVER_MODE")
	if !ok {
		t.Errorf("Expected FLOTILLA_SERVER_MODE to exist in environment")
	}

	if srvMode != "prod" {
		t.Errorf("Expected to pull server mode [%s], was [%s]", "prod", srvMode)
	}

	receipt.Done()
}
