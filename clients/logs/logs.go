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
func NewLogsClient(conf config.Config, logger flotillaLog.Logger) (Client, error) {
	name := "awslogs"
	if conf.IsSet("log.driver.name") {
		name = conf.GetString("log.driver.name")
	}

	logger.Log("message", "Initializing logs client", "client", name)
	switch name {
	case "awslogs":
		// awslogs as an ecs log driver sends logs to AWS CloudWatch Logs service
		cwlc := &CloudWatchLogsClient{}
		if err := cwlc.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing CloudWatchLogsClient")
		}
		return cwlc, nil
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
