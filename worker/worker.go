package worker

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Worker defines a background worker process
//
type Worker interface {
	Run()
}

func NewWorker(
	workerType string,
	log flotillaLog.Logger,
	conf config.Config,
	ee engine.Engine,
	sm state.Manager) (Worker, error) {
	switch workerType {
	case "submit":
		return &submitWorker{
			conf: conf,
			sm:   sm,
			ee:   ee,
			log:  log,
		}, nil
	case "retry":
		return &retryWorker{
			sm:   sm,
			ee:   ee,
			conf: conf,
			log:  log,
		}, nil
	case "status":
		return &statusWorker{
			sm:   sm,
			ee:   ee,
			conf: conf,
			log:  log,
		}, nil
	default:
		return nil, fmt.Errorf("No workerType %s exists", workerType)
	}
}
