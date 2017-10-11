package worker

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
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
	qm queue.Manager,
	sm state.Manager) (Worker, error) {
	switch workerType {
	case "submit":
		return &submitWorker{
			conf: conf,
			sm:   sm,
			qm:   qm,
			ee:   ee,
			log:  log,
		}, nil
	case "retry":
		return &retryWorker{
			sm:   sm,
			qm:   qm,
			conf: conf,
			log:  log,
		}, nil
	case "status":
		return &statusWorker{
			sm:   sm,
			qm:   qm,
			conf: conf,
			log:  log,
		}, nil
	default:
		return nil, fmt.Errorf("No workerType %s exists", workerType)
	}
}
