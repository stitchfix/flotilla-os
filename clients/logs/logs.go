package logs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Client returns logs for a Run
//
type Client interface {
	Name() string
	Initialize(config config.Config) error
	Logs(definition state.Definition, run state.Run, lastSeen *string) (string, *string, error)
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
// NewLogsClient creates and initializes a run logs client
//
func NewLogsClient(conf config.Config, logger flotillaLog.Logger, name string) (Client, error) {
	logger.Log("message", "Initializing logs client", "client", name)
	switch name {
	case "ecs":
		// awslogs as an ecs log driver sends logs to AWS CloudWatch Logs service
		cwlc := &ECSCloudWatchLogsClient{}
		if err := cwlc.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing ECSCloudWatchLogsClient")
		}
		return cwlc, nil
	case "eks":
		// awslogs as an ecs log driver sends logs to AWS CloudWatch Logs service
		ekscw := &EKSS3LogsClient{}
		if err := ekscw.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing EKSCloudWatchLogsClient")
		}
		return ekscw, nil
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
