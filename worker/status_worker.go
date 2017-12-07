package worker

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"time"
)

type statusWorker struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
}

func (sw *statusWorker) Initialize(
	conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = ee
	sw.log = log
	return nil
}

//
// Run updates status of tasks
//
func (sw *statusWorker) Run() {
	for {
		sw.runOnce()
		time.Sleep(sw.pollInterval)
	}
}

func (sw *statusWorker) runOnce() {
	runReceipt, err := sw.ee.PollStatus()
	if err != nil {
		sw.log.Log("message", "unable to receive status message", "error", err.Error())
		return
	}

	// Ensure update is in the env required, otherwise, ack without taking action
	update := runReceipt.Run
	if update != nil {
		//
		// Relies on the reserved env var, FLOTILLA_SERVER_MODE to ensure update
		// belongs to -this- mode of Flotilla
		//
		var serverMode string
		if update.Env != nil {
			for _, kv := range *update.Env {
				if kv.Name == "FLOTILLA_SERVER_MODE" {
					serverMode = kv.Value
				}
			}
		}

		shouldProcess := len(serverMode) > 0 && serverMode == sw.conf.GetString("flotilla_mode")
		if shouldProcess {
			run, err := sw.findRun(update.TaskArn)
			if err != nil {
				sw.log.Log("message", "unable to find run to apply update to", "error", err.Error())
				return
			}
			run.UpdateWith(*update)
			_, err = sw.sm.UpdateRun(run.RunID, run)
			if err != nil {
				sw.log.Log("message", "error applying status update", "run", run.RunID, "error", err.Error())
				return
			}
		}

		sw.log.Log("message", "Acking status update", "arn", update.TaskArn)
		if err = runReceipt.Done(); err != nil {
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
