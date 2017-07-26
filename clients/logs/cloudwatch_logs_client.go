package logs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stitchfix/flotilla-os/config"
	"sort"
	"strings"
)

//
// CloudWatchLogsClient corresponds with the aws logs driver
// for ECS and returns logs for runs
//
type CloudWatchLogsClient struct {
	logRetentionInDays int64
	logNamespace       string
	logsClient         logsClient
}

type logsClient interface {
	DescribeLogGroups(input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	CreateLogGroup(input *cloudwatchlogs.CreateLogGroupInput) (*cloudwatchlogs.CreateLogGroupOutput, error)
	PutRetentionPolicy(input *cloudwatchlogs.PutRetentionPolicyInput) (*cloudwatchlogs.PutRetentionPolicyOutput, error)
	GetLogEvents(input *cloudwatchlogs.GetLogEventsInput) (*cloudwatchlogs.GetLogEventsOutput, error)
}

type byTimestamp []*cloudwatchlogs.OutputLogEvent

func (events byTimestamp) Len() int           { return len(events) }
func (events byTimestamp) Swap(i, j int)      { events[i], events[j] = events[j], events[i] }
func (events byTimestamp) Less(i, j int) bool { return *(events[i].Timestamp) < *(events[j].Timestamp) }

//
// Name returns the name of the logs client
//
func (cwl *CloudWatchLogsClient) Name() string {
	return "cloudwatch"
}

//
// Initialize sets up the CloudWatchLogsClient
//
func (cwl *CloudWatchLogsClient) Initialize(conf config.Config) error {
	if !conf.IsSet("aws_default_region") {
		return fmt.Errorf("CloudWatchLogsClient needs [aws_default_region] set in config")
	}

	if !conf.IsSet("log.namespace") {
		return fmt.Errorf("CloudWatchLogsClient needs [log.namespace] set in config")
	}

	cwl.logNamespace = conf.GetString("log.namespace")
	cwl.logRetentionInDays = int64(conf.GetInt("log.retention_days"))
	if cwl.logRetentionInDays == 0 {
		cwl.logRetentionInDays = int64(30)
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(conf.GetString("aws_default_region"))}))

		cwl.logsClient = cloudwatchlogs.New(sess)
	}
	return cwl.createNamespaceIfNotExists()
}

//
// Logs returns all logs from the log stream identified by handle since lastSeen
//
func (cwl *CloudWatchLogsClient) Logs(handle string, lastSeen *string) (string, *string, error) {
	startFromHead := true
	args := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &cwl.logNamespace,
		LogStreamName: &handle,
		StartFromHead: &startFromHead,
	}

	if lastSeen != nil && len(*lastSeen) > 0 {
		args.NextToken = lastSeen
	}

	result, err := cwl.logsClient.GetLogEvents(args)
	if err != nil {
		return "", nil, err
	}

	if len(result.Events) == 0 {
		return "", result.NextForwardToken, nil
	}

	message := cwl.logsToMessage(result.Events)
	return message, result.NextForwardToken, nil
}

func (cwl *CloudWatchLogsClient) logsToMessage(events []*cloudwatchlogs.OutputLogEvent) string {
	sort.Sort(byTimestamp(events))

	messages := make([]string, len(events))
	for i, event := range events {
		messages[i] = *event.Message
	}
	return strings.Join(messages, "\n")
}

func (cwl *CloudWatchLogsClient) createNamespaceIfNotExists() error {
	exists, err := cwl.namespaceExists()
	if err != nil {
		return err
	}
	if !exists {
		return cwl.createNamespace()
	}
	return nil
}

func (cwl *CloudWatchLogsClient) namespaceExists() (bool, error) {
	result, err := cwl.logsClient.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &cwl.logNamespace,
	})

	if err != nil {
		return false, err
	}
	if len(result.LogGroups) == 0 {
		return false, nil
	}
	for _, group := range result.LogGroups {
		if *group.LogGroupName == cwl.logNamespace {
			return true, nil
		}
	}
	return false, nil
}

func (cwl *CloudWatchLogsClient) createNamespace() error {
	_, err := cwl.logsClient.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: &cwl.logNamespace,
	})
	if err != nil {
		return err
	}

	_, err = cwl.logsClient.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    &cwl.logNamespace,
		RetentionInDays: &cwl.logRetentionInDays,
	})
	return err
}
