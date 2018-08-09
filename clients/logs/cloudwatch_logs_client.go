package logs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
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
	logStreamPrefix    string
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
	confLogOptions := conf.GetStringMapString("log.driver.options")

	awsRegion := confLogOptions["awslogs-region"]
	if len(awsRegion) == 0 {
		awsRegion = conf.GetString("aws_default_region")
	}

	if len(awsRegion) == 0 {
		return errors.Errorf(
			"CloudWatchLogsClient needs one of [log.driver.options.awslogs-region] or [aws_default_region] set in config")
	}

	//
	// log.namespace in conf takes precedence over log.driver.options.awslogs-group
	//
	cwl.logNamespace = conf.GetString("log.namespace")
	if _, ok := confLogOptions["awslogs-group"]; ok && len(cwl.logNamespace) == 0 {
		cwl.logNamespace = confLogOptions["awslogs-group"]
	}

	if len(cwl.logNamespace) == 0 {
		return errors.Errorf(
			"CloudWatchLogsClient needs one of [log.driver.options.awslogs-group] or [log.namespace] set in config")
	}

	cwl.logStreamPrefix = confLogOptions["awslogs-stream-prefix"]
	if len(cwl.logStreamPrefix) == 0 {
		return errors.Errorf("CloudWatchLogsClient needs [log.driver.options.awslogs-stream-prefix] set in config")
	}

	cwl.logRetentionInDays = int64(conf.GetInt("log.retention_days"))
	if cwl.logRetentionInDays == 0 {
		cwl.logRetentionInDays = int64(30)
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)}))

		cwl.logsClient = cloudwatchlogs.New(sess)
	}
	return cwl.createNamespaceIfNotExists()
}

//
// Logs returns all logs from the log stream identified by handle since lastSeen
//
func (cwl *CloudWatchLogsClient) Logs(definition state.Definition, run state.Run, lastSeen *string) (string, *string, error) {
	startFromHead := true
	handle := cwl.toStreamName(definition, run)
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
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == cloudwatchlogs.ErrCodeResourceNotFoundException {
				// Fallback logic for legacy container names
				if strings.HasPrefix(definition.ContainerName, definition.GroupName) {
					definition.ContainerName = strings.Replace(
						definition.ContainerName, fmt.Sprintf("%s-", definition.GroupName), "", -1)
					return cwl.Logs(definition, run, lastSeen)
				}

				return "", nil, exceptions.MissingResource{err.Error()}
			}
		}
		return "", nil, errors.Wrap(err, "problem getting logs")
	}

	if len(result.Events) == 0 {
		return "", result.NextForwardToken, nil
	}

	message := cwl.logsToMessage(result.Events)
	return message, result.NextForwardToken, nil
}

func (cwl *CloudWatchLogsClient) toStreamName(definition state.Definition, run state.Run) string {
	arnSplits := strings.Split(run.TaskArn, "/")
	return fmt.Sprintf(
		"%s/%s/%s", cwl.logStreamPrefix, definition.ContainerName, arnSplits[len(arnSplits)-1])
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
		return errors.Wrapf(err, "problem checking if log namespace [%s] exists", cwl.logNamespace)
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
		return false, errors.Wrapf(err, "problem describing log groups with prefix [%s]", cwl.logNamespace)
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
		return errors.Wrapf(err, "problem creating log group with log group name [%s]", cwl.logNamespace)
	}

	_, err = cwl.logsClient.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    &cwl.logNamespace,
		RetentionInDays: &cwl.logRetentionInDays,
	})
	if err != nil {
		return errors.Wrapf(err, "problem setting log group retention policy for log group name [%s]", cwl.logNamespace)
	}
	return nil
}
