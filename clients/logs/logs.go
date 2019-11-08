package logs

import (
	"fmt"
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
		//TODO
		return nil, errors.New("TODO - NOT IMPLEMENTED")
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
