package logs

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
	"log"
	"os"
	"sort"
	"strings"
)

//
// EKSCloudWatchLogsClient corresponds with the aws logs driver
// for ECS and returns logs for runs
//
type EKSCloudWatchLogsClient struct {
	logRetentionInDays int64
	logNamespace       string
	logsClient         logsClient
	logger             *log.Logger
}

type EKSCloudWatchLog struct {
	Log string `json:"log"`
}

//
// Name returns the name of the logs client
//
func (lc *EKSCloudWatchLogsClient) Name() string {
	return "eks-cloudwatch"
}

//
// Initialize sets up the EKSCloudWatchLogsClient
//
func (lc *EKSCloudWatchLogsClient) Initialize(conf config.Config) error {
	confLogOptions := conf.GetStringMapString("eks.log.driver.options")

	awsRegion := confLogOptions["awslogs-region"]
	if len(awsRegion) == 0 {
		awsRegion = conf.GetString("aws_default_region")
	}

	if len(awsRegion) == 0 {
		return errors.Errorf(
			"EKSCloudWatchLogsClient needs one of [eks.log.driver.options.awslogs-region] or [aws_default_region] set in config")
	}

	//
	// log.namespace in conf takes precedence over log.driver.options.awslogs-group
	//
	lc.logNamespace = conf.GetString("eks.log.namespace")
	if _, ok := confLogOptions["awslogs-group"]; ok && len(lc.logNamespace) == 0 {
		lc.logNamespace = confLogOptions["awslogs-group"]
	}

	if len(lc.logNamespace) == 0 {
		return errors.Errorf(
			"EKSCloudWatchLogsClient needs one of [eks.log.driver.options.awslogs-group] or [eks.log.namespace] set in config")
	}

	lc.logRetentionInDays = int64(conf.GetInt("eks.log.retention_days"))
	if lc.logRetentionInDays == 0 {
		lc.logRetentionInDays = int64(30)
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := session.Must(session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)}))

		lc.logsClient = cloudwatchlogs.New(sess)
	}
	lc.logger = log.New(os.Stderr, "[cloudwatchlogs] ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return lc.createNamespaceIfNotExists()
}

//
// Logs returns all logs from the log stream identified by handle since lastSeen
//
func (lc *EKSCloudWatchLogsClient) Logs(definition state.Definition, run state.Run, lastSeen *string) (string, *string, error) {
	startFromHead := true

	if run.PodName == nil {
		return "", nil, errors.New("problem getting logs")
	}
	handle := lc.toStreamName(run)
	args := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  &lc.logNamespace,
		LogStreamName: &handle,
		StartFromHead: &startFromHead,
	}

	if lastSeen != nil && len(*lastSeen) > 0 {
		args.NextToken = lastSeen
	}

	result, err := lc.logsClient.GetLogEvents(args)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == cloudwatchlogs.ErrCodeResourceNotFoundException {
				// Fallback logic for legacy container names
				if strings.HasPrefix(definition.ContainerName, definition.GroupName) {
					definition.ContainerName = strings.Replace(
						definition.ContainerName, fmt.Sprintf("%s-", definition.GroupName), "", -1)
					return lc.Logs(definition, run, lastSeen)
				}

				return "", nil, exceptions.MissingResource{err.Error()}
			} else if request.IsErrorThrottle(err) {
				lc.logger.Printf(
					"thottled getting logs; definition_id: %s, run_id: %s, error: %+v\n",
					definition.DefinitionID, run.RunID, err)
				return "", lastSeen, nil
			}
		}
		return "", nil, errors.Wrap(err, "problem getting logs")
	}

	if len(result.Events) == 0 {
		return "", result.NextForwardToken, nil
	}

	message := lc.logsToMessage(result.Events)
	return message, result.NextForwardToken, nil
}

func (lc *EKSCloudWatchLogsClient) toStreamName(run state.Run) string {

	return fmt.Sprintf("%s", *run.PodName)
}

func (lc *EKSCloudWatchLogsClient) logsToMessage(events []*cloudwatchlogs.OutputLogEvent) string {
	sort.Sort(byTimestamp(events))

	messages := make([]string, len(events))
	for i, event := range events {
		var l EKSCloudWatchLog
		err := json.Unmarshal([]byte(*event.Message), &l)
		if err != nil {
			messages[i] = *event.Message
		}
		messages[i] = l.Log
	}
	return strings.Join(messages, "")
}

func (lc *EKSCloudWatchLogsClient) createNamespaceIfNotExists() error {
	exists, err := lc.namespaceExists()
	if err != nil {
		return errors.Wrapf(err, "problem checking if log namespace [%s] exists", lc.logNamespace)
	}
	if !exists {
		return lc.createNamespace()
	}
	return nil
}

func (lc *EKSCloudWatchLogsClient) namespaceExists() (bool, error) {
	result, err := lc.logsClient.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &lc.logNamespace,
	})

	if err != nil {
		return false, errors.Wrapf(err, "problem describing log groups with prefix [%s]", lc.logNamespace)
	}
	if len(result.LogGroups) == 0 {
		return false, nil
	}
	for _, group := range result.LogGroups {
		if *group.LogGroupName == lc.logNamespace {
			return true, nil
		}
	}
	return false, nil
}

func (lc *EKSCloudWatchLogsClient) createNamespace() error {
	_, err := lc.logsClient.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: &lc.logNamespace,
	})
	if err != nil {
		return errors.Wrapf(err, "problem creating log group with log group name [%s]", lc.logNamespace)
	}

	_, err = lc.logsClient.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
		LogGroupName:    &lc.logNamespace,
		RetentionInDays: &lc.logRetentionInDays,
	})
	if err != nil {
		return errors.Wrapf(err, "problem setting log group retention policy for log group name [%s]", lc.logNamespace)
	}
	return nil
}
