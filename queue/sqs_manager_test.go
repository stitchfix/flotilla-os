package queue

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"testing"
)

type testSQSClient struct {
	t      *testing.T
	queues []*string
}

func (qc *testSQSClient) ListQueues(input *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error) {
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
	jsonRun, _ := json.Marshal(state.Run{RunID: "cupcake"})
	asString := string(jsonRun)

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

	qA := "A"
	qB := "B"
	qC := "C"
	testClient := testSQSClient{
		t:      t,
		queues: []*string{&qA, &qB, &qC},
	}
	qm.qc = &testClient

	return qm
}

func TestSQSManager_List(t *testing.T) {
	qm := setUp(t)

	listed, _ := qm.List()
	if len(listed) != 3 {
		t.Errorf("Expected listed queues to be [3] but was %v", len(listed))
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

func TestSQSManager_Receive(t *testing.T) {
	qm := setUp(t)
	receipt, _ := qm.Receive("A")
	receipt.Done()
}
