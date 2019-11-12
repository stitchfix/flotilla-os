package worker

import (
	"fmt"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
)

type retryWorker struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	engine       *string
}

func (rw *retryWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string) error {
	rw.pollInterval = pollInterval
	rw.conf = conf
	rw.sm = sm
	rw.ee = ee
	rw.log = log
	rw.engine = engine
	rw.log.Log("message", "initialized a retry worker", "engine", *engine)

	rw.engine = engine

	return nil
}

func (rw *retryWorker) GetTomb() *tomb.Tomb {
	return &rw.t
}

//
// Run finds tasks that NEED_RETRY and requeues them
//
func (rw *retryWorker) Run() error {
	for {
		select {
		case <-rw.t.Dying():
			rw.log.Log("message", "A retry worker was terminated")
			return nil
		default:
			rw.runOnce()
			time.Sleep(rw.pollInterval)
		}
	}
}

func (rw *retryWorker) runOnce() {
	// List runs in the StatusNeedsRetry state and requeue them
	var engines []string
	if rw.engine != nil {
		engines = []string{*rw.engine}
	} else {
		engines = nil
	}

	runList, err := rw.sm.ListRuns(25, 0, "started_at", "asc", map[string][]string{"status": {state.StatusNeedsRetry}}, nil, engines)

	rw.log.Log("message", fmt.Sprintf("Got %v jobs to retry", runList.Total))

	if err != nil {
		rw.log.Log("message", "Error listing runs for retry", "error", fmt.Sprintf("%+v", err))
		return
	}

	for _, run := range runList.Runs {

		if _, err = rw.sm.UpdateRun(run.RunID, state.Run{Status: state.StatusQueued}); err != nil {
			rw.log.Log("message", "Error updating run status to StatusQueued", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
			return
		}

		if err = rw.ee.Enqueue(run); err != nil {
			rw.log.Log("message", "Error enqueuing run", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
			return
		}
	}
	return
}
