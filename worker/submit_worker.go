package worker

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"time"
)

type submitWorker struct {
	sm   state.Manager
	qm   queue.Manager
	ee   engine.Engine
	conf config.Config
	log  flotillaLog.Logger
}

//
// Run lists queues, consumes runs from them, and executes them using the execution engine
//
func (sw *submitWorker) Run() {
	pollIntervalSeconds := sw.conf.GetInt("worker.submit_interval_seconds")
	if pollIntervalSeconds == 0 {
		pollIntervalSeconds = 30
	}
	pollInterval := time.Duration(pollIntervalSeconds) * time.Second
	for {
		sw.runOnce()
		time.Sleep(pollInterval)
	}
}

func (sw *submitWorker) runOnce() {
	queues, err := sw.qm.List()
	if err != nil {
		sw.log.Log("message", "Error listing queues", "error", err.Error())
	}

	for _, qurl := range queues {

		//
		// Get new queued Run
		//
		runReceipt, err := sw.qm.ReceiveRun(qurl)
		if err != nil {
			sw.log.Log("message", "Error receiving run", "qurl", qurl, "error", err.Error())
			continue
		}

		if runReceipt.Run == nil {
			continue
		}

		//
		// Fetch run from state manager to ensure its existence
		//
		run, err := sw.sm.GetRun(runReceipt.Run.RunID)
		if err != nil {
			sw.log.Log("message", "Error fetching run from state, acking", "run_id", runReceipt.Run.RunID, "error", err.Error())
			if err = runReceipt.Done(); err != nil {
				sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", err.Error())
			}
			continue
		}

		//
		// Fetch run's definition from state manager
		//
		// * Will not be necessary once we copy relevant run information from definition onto the run itself
		//
		definition, err := sw.sm.GetDefinition(run.DefinitionID)
		if err != nil {
			sw.log.Log(
				"message", "Error fetching definition for run",
				"run_id", run.RunID,
				"definition_id", run.DefinitionID,
				"error", err.Error())
			if err = runReceipt.Done(); err != nil {
				sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", err.Error())
			}
			continue
		}

		//
		// Only valid to process if it's in the StatusQueued state
		//
		if run.Status == state.StatusQueued {

			//
			// Execute the run using the execution engine
			//
			sw.log.Log("message", "Submitting", "run_id", run.RunID)
			launched, retryable, err := sw.ee.Execute(definition, run)
			if err != nil {
				sw.log.Log("message", "Error executing run", "run_id", run.RunID, "error", err.Error(), "retryable", retryable)
				if !retryable {
					// Set status to StatusStopped, and ack
					launched.Status = state.StatusStopped
				} else {
					// Don't change status, don't ack
					continue
				}
			}

			//
			// Emit event with current definition
			//
			err = sw.log.Event("eventClassName", "FlotillaSubmitTask", "definition", definition, "run_id", run.RunID)
			if err != nil {
				sw.log.Log("message", "Failed to emit event", "run_id", run.RunID, "error", err.Error()) }

			//
			// Update the status and information of the run;
			// either the run submitted successfully -or- it did not and is not retryable
			//
			if _, err = sw.sm.UpdateRun(run.RunID, launched); err != nil {
				sw.log.Log("message", "Failed to update run status", "run_id", run.RunID, "status", launched.Status, "error", err.Error())
			}
		} else {
			sw.log.Log("message", "Received run that is not runnable", "run_id", run.RunID, "status", run.Status)
		}

		if err = runReceipt.Done(); err != nil {
			sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", err.Error())
		}
	}
}
