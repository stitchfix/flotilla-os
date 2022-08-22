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
	"net/http"
	"os"
	"sort"
	"strings"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
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
	//confLogOptions := conf.GetStringMapString("eks_log_driver_options")

	awsRegion := conf.GetString("eks_log_driver_options_awslogs_region")
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
	lc.logNamespace = conf.GetString("eks_log_namespace")

	if len(lc.logNamespace) == 0 {
		return errors.Errorf(
			"EKSCloudWatchLogsClient needs one of [eks.log.driver.options.awslogs-group] or [eks.log.namespace] set in config")
	}

	lc.logRetentionInDays = int64(conf.GetInt("eks_log_retention_days"))
	if lc.logRetentionInDays == 0 {
		lc.logRetentionInDays = int64(30)
	}

	flotillaMode := conf.GetString("flotilla_mode")
	if flotillaMode != "test" {
		sess := awstrace.WrapSession(session.Must(session.NewSession(&aws.Config{
			Region: aws.String(awsRegion)})))

		lc.logsClient = cloudwatchlogs.New(sess)
	}
	lc.logger = log.New(os.Stderr, "[cloudwatchlogs] ",
		log.Ldate|log.Ltime|log.Lshortfile)
	return lc.createNamespaceIfNotExists()
}

//
// Logs returns all logs from the log stream identified by handle since lastSeen
//
func (lc *EKSCloudWatchLogsClient) Logs(executable state.Executable, run state.Run, lastSeen *string, role *string, facility *string) (string, *string, error) {
	startFromHead := true

	//Pod isn't there yet - dont return a 404
	if run.PodName == nil {
		return "", nil, nil
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
				return "", nil, exceptions.MissingResource{err.Error()}
			} else if request.IsErrorThrottle(err) {
				lc.logger.Printf(
					"thottled getting logs; executable_id: %v, run_id: %s, error: %+v\n",
					executable.GetExecutableID(), run.RunID, err)
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

// This method doesn't return log string, it is a placeholder only.
func (lc *EKSCloudWatchLogsClient) LogsText(executable state.Executable, run state.Run, w http.ResponseWriter) error {
	return errors.Errorf("EKSCloudWatchLogsClient does not support LogsText method.")
}

// Generate stream name
func (lc *EKSCloudWatchLogsClient) toStreamName(run state.Run) string {
	return fmt.Sprintf("%s", *run.PodName)
}

// Convert Cloudwatch logs to strings
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

// Check for the existence of a namespace.
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

// Creates namespace is not present.
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
