package worker

import (
	"fmt"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

type retryWorker struct {
	sm  state.Manager
	qm  queue.Manager
	log flotillaLog.Logger
}

func (rw *retryWorker) Run() {
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
		qurl, err := rw.qm.QurlFor(run.ClusterName)

		if err != nil {
			rw.log.Log("message", "Error getting QurlFor cluster", "cluster", run.ClusterName, "error", err.Error())
			return
		}

		if err = rw.sm.UpdateRun(run.RunID, state.Run{Status: state.StatusQueued}); err != nil {
			rw.log.Log("message", "Error updating run status to StatusQueued", "run_id", run.RunID, "error", err.Error())
			return
		}

		if err = rw.qm.Enqueue(qurl, run); err != nil {
			rw.log.Log("message", "Error enqueuing run", "run_id", run.RunID, "qurl", qurl, "error", err.Error())
			return
		}
	}
	return
}
