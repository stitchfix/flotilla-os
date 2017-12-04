package worker

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"time"
)

type retryWorker struct {
	sm   state.Manager
	ee   engine.Engine
	conf config.Config
	log  flotillaLog.Logger
}

//
// Run finds tasks that NEED_RETRY and requeues them
//
func (rw *retryWorker) Run() {
	pollIntervalSeconds := rw.conf.GetInt("worker.retry_interval_seconds")
	if pollIntervalSeconds == 0 {
		pollIntervalSeconds = 30
	}
	pollInterval := time.Duration(pollIntervalSeconds) * time.Second
	for {
		rw.runOnce()
		time.Sleep(pollInterval)
	}
}

func (rw *retryWorker) runOnce() {
	// List runs in the StatusNeedsRetry state and requeue them
	runList, err := rw.sm.ListRuns(
		25, 0,
		"started_at", "asc",
		map[string]string{"status": state.StatusNeedsRetry}, nil)

	rw.log.Log("message", fmt.Sprintf("Got %v jobs to retry", runList.Total))

	if err != nil {
		rw.log.Log("message", "Error listing runs for retry", "error", err.Error())
		return
	}

	for _, run := range runList.Runs {

		if _, err = rw.sm.UpdateRun(run.RunID, state.Run{Status: state.StatusQueued}); err != nil {
			rw.log.Log("message", "Error updating run status to StatusQueued", "run_id", run.RunID, "error", err.Error())
			return
		}

		if err = rw.ee.Enqueue(run); err != nil {
			rw.log.Log("message", "Error enqueuing run", "run_id", run.RunID, "error", err.Error())
			return
		}
	}
	return
}
