package logs

import (
	"fmt"
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
	name := conf.GetString("log.driver.name")
	if len(name) == 0 {
		name = "awslogs"
	}

	logger.Log("message", "Initializing logs client", "client", name)
	switch name {
	case "awslogs":
		// awslogs as an ecs log driver sends logs to AWS CloudWatch Logs service
		cwlc := &CloudWatchLogsClient{}
		return cwlc, cwlc.Initialize(conf)
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
