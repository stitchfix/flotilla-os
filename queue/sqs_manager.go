package queue

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
)

//
// SQSManager - queue manager implementation for sqs
//
type SQSManager struct {
	namespace         string
	retentionSeconds  string
	visibilityTimeout string
	qc                sqsClient
	qurlCache         map[string]string
}

type sqsClient interface {
	GetQueueUrl(input *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error)
	CreateQueue(input *sqs.CreateQueueInput) (*sqs.CreateQueueOutput, error)
	ListQueues(input *sqs.ListQueuesInput) (*sqs.ListQueuesOutput, error)
	SendMessage(input *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
	ReceiveMessage(input *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
}

//
// Name of queue manager - matches value in configuration
//
func (qm *SQSManager) Name() string {
	return "sqs"
}

//
// Initialize new sqs queue manager
//
func (qm *SQSManager) Initialize(conf config.Config, engine string) error {
	if !conf.IsSet("aws_default_region") {
		return errors.Errorf("SQSManager needs [aws_default_region] set in config")
	}

	qm.retentionSeconds = "604800"
	if conf.IsSet("queue_retention_seconds") {
		qm.retentionSeconds = conf.GetString("queue_retention_seconds")
	}

	qm.visibilityTimeout = "45"
	if conf.IsSet("queue_process_time") {
		qm.visibilityTimeout = conf.GetString("queue_process_time")
	}

	if !conf.IsSet("queue_namespace") {
		return errors.Errorf("SQSManager needs [queue_namespace] set in config")
	}

	qm.namespace = conf.GetString("queue_namespace")
	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := awstrace.WrapSession(session.Must(session.NewSession(&aws.Config{
			Region: aws.String(conf.GetString("aws_default_region"))})))

		qm.qc = sqs.New(sess)
	}

	qm.qurlCache = make(map[string]string)
	return nil
}

//
// QurlFor returns the queue url that corresponds to the given name
// * if the queue does not exist it is created
//
func (qm *SQSManager) QurlFor(name string, prefixed bool) (string, error) {
	key := fmt.Sprintf("%s-%t", name, prefixed)
	val, ok := qm.qurlCache[key]
	if ok {
		return val, nil
	}

	val, err := qm.getOrCreateQueue(name, prefixed)
	if err == nil {
		qm.qurlCache[key] = val
	}
	return val, err
}

func (qm *SQSManager) getOrCreateQueue(name string, prefixed bool) (string, error) {
	qname := name
	if prefixed {
		qname = fmt.Sprintf("%s-%s", qm.namespace, name)
	}
	res, err := qm.qc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qname,
	})
	if err != nil || res.QueueUrl == nil {
		cqi := sqs.CreateQueueInput{
			Attributes: map[string]*string{
				"MessageRetentionPeriod": &qm.retentionSeconds,
				"VisibilityTimeout":      &qm.visibilityTimeout,
			},
			QueueName: &qname,
		}
		createQueueResponse, err := qm.qc.CreateQueue(&cqi)
		if err != nil {
			return "", errors.Wrapf(err, "problem trying to create sqs queue with name [%s]", qname)
		}
		return *createQueueResponse.QueueUrl, nil
	}
	return *res.QueueUrl, nil
}

func (qm *SQSManager) messageFromRun(run state.Run) (*string, error) {
	jsonized, err := json.Marshal(run)
	if err != nil {
		return nil, errors.Wrapf(err, "problem trying to serialize run with id [%s] as json", run.RunID)
	}
	asString := string(jsonized)
	return &asString, nil
}

func (qm *SQSManager) runFromMessage(message *sqs.Message) (state.Run, error) {
	var run state.Run
	if message == nil {
		return run, errors.Errorf("can't generate Run from nil message")
	}

	body := message.Body
	if body == nil {
		return run, errors.Errorf("can't generate Run from empty message")
	}

	if err := json.Unmarshal([]byte(*body), &run); err != nil {
		errors.Wrapf(err, "problem trying to deserialize run from json [%s]", *body)
	}

	return run, nil
}

func (qm *SQSManager) statusFromMessage(message *sqs.Message) (string, error) {
	var statusUpdate string
	if message == nil {
		return statusUpdate, errors.Errorf("can't generate StatusUpdate from nil message")
	}

	body := message.Body
	if body == nil {
		return statusUpdate, errors.Errorf("can't generate StatusUpdate from empty message")
	}

	return *body, nil
}

//
// Enqueue queues run
//
func (qm *SQSManager) Enqueue(qURL string, run state.Run) error {
	if len(qURL) == 0 {
		return errors.Errorf("no queue url specified, can't enqueue")
	}

	message, err := qm.messageFromRun(run)
	if err != nil {
		return errors.WithStack(err)
	}

	sme := sqs.SendMessageInput{
		QueueUrl:    &qURL,
		MessageBody: message,
	}

	_, err = qm.qc.SendMessage(&sme)
	if err != nil {
		return errors.Wrap(err, "problem sending sqs message")
	}
	return nil
}

//
// Receive receives a new run to operate on
//
func (qm *SQSManager) ReceiveRun(qURL string) (RunReceipt, error) {
	var receipt RunReceipt

	if len(qURL) == 0 {
		return receipt, errors.Errorf("no queue url specified, can't dequeue")
	}

	maxMessages := int64(1)
	visibilityTimeout := int64(45)
	rmi := sqs.ReceiveMessageInput{
		QueueUrl:            &qURL,
		MaxNumberOfMessages: &maxMessages,
		VisibilityTimeout:   &visibilityTimeout,
	}

	var err error

	response, err := qm.qc.ReceiveMessage(&rmi)
	if err != nil {
		return receipt, errors.Wrapf(err, "problem receiving sqs message from queue url [%s]", qURL)
	}

	if len(response.Messages) == 0 {
		return receipt, nil
	}

	run, err := qm.runFromMessage(response.Messages[0])
	if err != nil {
		return receipt, errors.WithStack(err)
	}

	receipt.Run = &run
	receipt.Done = func() error {
		return qm.ack(qURL, response.Messages[0].ReceiptHandle)
	}
	return receipt, nil
}

func (qm *SQSManager) ReceiveStatus(qURL string) (StatusReceipt, error) {
	var receipt StatusReceipt

	if len(qURL) == 0 {
		return receipt, errors.Errorf("no queue url specified, can't dequeue")
	}

	maxMessages := int64(1)
	visibilityTimeout := int64(45)
	rmi := sqs.ReceiveMessageInput{
		QueueUrl:            &qURL,
		MaxNumberOfMessages: &maxMessages,
		VisibilityTimeout:   &visibilityTimeout,
	}

	var err error

	response, err := qm.qc.ReceiveMessage(&rmi)
	if err != nil {
		return receipt, errors.Wrapf(err, "problem receiving sqs message from queue url [%s]", qURL)
	}

	if len(response.Messages) == 0 {
		return receipt, nil
	}

	statusUpdate, err := qm.statusFromMessage(response.Messages[0])
	if err != nil {
		return receipt, errors.WithStack(err)
	}
	receipt.StatusUpdate = &statusUpdate
	receipt.Done = func() error {
		return qm.ack(qURL, response.Messages[0].ReceiptHandle)
	}
	return receipt, nil
}

func (qm *SQSManager) ReceiveCloudTrail(qURL string) (state.CloudTrailS3File, error) {
	var receipt state.CloudTrailS3File

	if len(qURL) == 0 {
		return receipt, errors.Errorf("no queue url specified, can't dequeue")
	}

	maxMessages := int64(1)
	visibilityTimeout := int64(45)
	rmi := sqs.ReceiveMessageInput{
		QueueUrl:            &qURL,
		MaxNumberOfMessages: &maxMessages,
		VisibilityTimeout:   &visibilityTimeout,
	}

	var err error

	response, err := qm.qc.ReceiveMessage(&rmi)
	if err != nil {
		return receipt, errors.Wrapf(err, "problem receiving sqs message from queue url [%s]", qURL)
	}

	if response != nil && response.Messages != nil && len(response.Messages) > 0 && response.Messages[0].Body != nil {
		body := response.Messages[0].Body

		err = json.Unmarshal([]byte(*body), &receipt)
		_ = qm.ack(qURL, response.Messages[0].ReceiptHandle)

	}
	return receipt, nil
}

func (qm *SQSManager) ReceiveEMREvent(qURL string) (state.EmrEvent, error) {
	var emrEvent state.EmrEvent

	if len(qURL) == 0 {
		return emrEvent, errors.Errorf("no queue url specified, can't dequeue")
	}

	maxMessages := int64(1)
	visibilityTimeout := int64(45)
	rmi := sqs.ReceiveMessageInput{
		QueueUrl:            &qURL,
		MaxNumberOfMessages: &maxMessages,
		VisibilityTimeout:   &visibilityTimeout,
	}

	var err error

	response, err := qm.qc.ReceiveMessage(&rmi)
	if err != nil {
		return emrEvent, errors.Wrapf(err, "problem receiving sqs message from queue url [%s]", qURL)
	}

	if response != nil && response.Messages != nil && len(response.Messages) > 0 && response.Messages[0].Body != nil {
		body := response.Messages[0].Body

		err = json.Unmarshal([]byte(*body), &emrEvent)
		emrEvent.Done = func() error {
			return qm.ack(qURL, response.Messages[0].ReceiptHandle)
		}

	}
	return emrEvent, nil
}

func (qm *SQSManager) ReceiveKubernetesEvent(qURL string) (state.KubernetesEvent, error) {
	var kubernetesEvent state.KubernetesEvent

	if len(qURL) == 0 {
		return kubernetesEvent, errors.Errorf("no queue url specified, can't dequeue")
	}

	maxMessages := int64(1)
	visibilityTimeout := int64(45)
	rmi := sqs.ReceiveMessageInput{
		QueueUrl:            &qURL,
		MaxNumberOfMessages: &maxMessages,
		VisibilityTimeout:   &visibilityTimeout,
	}

	var err error

	response, err := qm.qc.ReceiveMessage(&rmi)
	if err != nil {
		return kubernetesEvent, errors.Wrapf(err, "problem receiving sqs message from queue url [%s]", qURL)
	}

	if response != nil && response.Messages != nil && len(response.Messages) > 0 && response.Messages[0].Body != nil {
		body := response.Messages[0].Body

		err = json.Unmarshal([]byte(*body), &kubernetesEvent)
		kubernetesEvent.Done = func() error {
			return qm.ack(qURL, response.Messages[0].ReceiptHandle)
		}

	}
	return kubernetesEvent, nil
}

func (qm *SQSManager) ReceiveKubernetesRun(queue string) (string, error) {
	var runId string

	qURL, err := qm.QurlFor(queue, false)
	if len(qURL) == 0 || err != nil {
		return runId, errors.Errorf("no queue url specified, can't dequeue")
	}

	maxMessages := int64(1)
	visibilityTimeout := int64(45)
	rmi := sqs.ReceiveMessageInput{
		QueueUrl:            &qURL,
		MaxNumberOfMessages: &maxMessages,
		VisibilityTimeout:   &visibilityTimeout,
	}

	response, err := qm.qc.ReceiveMessage(&rmi)
	if err != nil {
		return runId, errors.Wrapf(err, "problem receiving sqs message from queue url [%s]", qURL)
	}

	if response != nil && response.Messages != nil && len(response.Messages) > 0 && response.Messages[0].Body != nil {
		_ = qm.ack(qURL, response.Messages[0].ReceiptHandle)
		return *response.Messages[0].Body, nil
	}

	return runId, errors.Wrapf(err, "no message")
}

//
// Ack acknowledges the receipt -AND- processing of the
// the message referred to by handle
//
func (qm *SQSManager) ack(qURL string, handle *string) error {
	if handle == nil {
		return errors.Errorf("cannot acknowledge message with nil receipt")
	}
	if len(*handle) == 0 {
		return errors.Errorf("cannot acknowledge message with empty receipt")
	}
	dmi := sqs.DeleteMessageInput{
		QueueUrl:      &qURL,
		ReceiptHandle: handle,
	}
	if _, err := qm.qc.DeleteMessage(&dmi); err != nil {
		return errors.Wrapf(
			err, "problem deleting sqs message with handle [%s] from queue url [%s]", *handle, qURL)
	}
	return nil
}

//
// List lists all the queue URLS available
//
func (qm *SQSManager) List() ([]string, error) {
	response, err := qm.qc.ListQueues(
		&sqs.ListQueuesInput{QueueNamePrefix: &qm.namespace})
	if err != nil {
		return nil, errors.Wrap(err, "problem listing sqs queues")
	}

	listed := make([]string, len(response.QueueUrls))
	for i, qurl := range response.QueueUrls {
		listed[i] = *qurl
	}
	return listed, nil
}
