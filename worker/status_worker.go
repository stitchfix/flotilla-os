package worker

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"time"
)

type statusWorker struct {
	sm   state.Manager
	qm   queue.Manager
	conf config.Config
	log  flotillaLog.Logger
}

//
// Run updates status of tasks
//
func (sw *statusWorker) Run() {
	pollIntervalSeconds := sw.conf.GetInt("worker.status_interval_seconds")
	if pollIntervalSeconds == 0 {
		pollIntervalSeconds = 5
	}
	pollInterval := time.Duration(pollIntervalSeconds) * time.Second

	statusQueue := sw.conf.GetString("queue.status")
	sw.log.Log("message", fmt.Sprintf("using status queue [%s]", statusQueue))

	qurl, err := sw.qm.QurlFor(statusQueue, false)
	if err != nil {
		sw.log.Log("message", "unable to start status worker, no qurl found", "error", err.Error())
	} else {
		for {
			sw.runOnce(qurl)
			time.Sleep(pollInterval)
		}
	}
}

func (sw *statusWorker) runOnce(statusQurl string) {
	statusReceipt, err := sw.qm.ReceiveStatus(statusQurl)
	if err != nil {
		sw.log.Log("message", "unable to receive status message", "error", err.Error())
		return
	}

	// Ensure update is in the env required, otherwise, ack without taking action
	update := statusReceipt.StatusUpdate
	if update != nil {
		//
		// Relies on the reserved env var, FLOTILLA_SERVER_MODE to ensure update
		// belongs to -this- mode of Flotilla
		//
		serverMode, ok := update.GetEnvVar("FLOTILLA_SERVER_MODE")
		shouldProcess := ok && serverMode == sw.conf.GetString("flotilla_mode")
		if shouldProcess {
			run, err := sw.findRun(update.TaskArn)
			if err != nil {
				sw.log.Log("message", "unable to find run to apply update to", "error", err.Error())
				return
			}
			run.UpdateStatus(update)
			_, err = sw.sm.UpdateRun(run.RunID, run)
			if err != nil {
				sw.log.Log("message", "error applying status update", "run", run.RunID, "error", err.Error())
				return
			}
		}

		sw.log.Log("message", "Acking status update", "arn", update.TaskArn)
		if err = statusReceipt.Done(); err != nil {
			sw.log.Log("message", "Acking status update failed", "arn", update.TaskArn, "error", err.Error())
		}
	}
}

func (sw *statusWorker) findRun(taskArn string) (state.Run, error) {
	runs, err := sw.sm.ListRuns(1, 0, "started_at", "asc", map[string]string{
		"task_arn": taskArn,
	}, nil)
	if err != nil {
		return state.Run{}, err
	}
	if runs.Total > 0 {
		return runs.Runs[0], nil
	}
	return state.Run{}, fmt.Errorf("No run found for [%s]", taskArn)
}
