package logs

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
)

//
// Client returns logs for a Run
//
type Client interface {
	Name() string
	Initialize(config config.Config) error
	Logs(handle string, lastSeen *string) (string, *string, error)
}

//
// NewLogsClient creates and initializes a run logs client
//
func NewLogsClient(conf config.Config) (Client, error) {
	name := conf.GetString("logs_client")
	if len(name) == 0 {
		name = "cloudwatch"
	}

	switch name {
	case "cloudwatch":
		cwlc := &CloudWatchLogsClient{}
		return cwlc, cwlc.Initialize(conf)
	default:
		return nil, fmt.Errorf("No Client named [%s] was found", name)
	}
}
