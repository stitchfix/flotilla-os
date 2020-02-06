package worker

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
	"time"
)

type submitWorker struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	engine       *string
	redisClient  *redis.Client
}

func (sw *submitWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string, qm queue.Manager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = ee
	sw.log = log
	sw.engine = engine
	sw.redisClient = redis.NewClient(&redis.Options{Addr: conf.GetString("redis_address"), DB: conf.GetInt("redis_db")})
	_ = sw.log.Log("message", "initialized a submit worker", "engine", *engine)
	return nil
}

func (sw *submitWorker) GetTomb() *tomb.Tomb {
	return &sw.t
}

//
// Run lists queues, consumes runs from them, and executes them using the execution engine
//
func (sw *submitWorker) Run() error {
	for {
		select {
		case <-sw.t.Dying():
			sw.log.Log("message", "A submit worker was terminated")
			return nil
		default:
			sw.runOnce()
			time.Sleep(sw.pollInterval)
		}
	}
}

func (sw *submitWorker) runOnce() {
	receipts, err := sw.ee.PollRuns()
	if err != nil {
		sw.log.Log("message", "Error receiving runs", "error", fmt.Sprintf("%+v", err))
	}
	for _, runReceipt := range receipts {
		if runReceipt.Run == nil {
			continue
		}

		//
		// Fetch run from state manager to ensure its existence
		//
		run, err := sw.sm.GetRun(runReceipt.Run.RunID)
		if err != nil {
			sw.log.Log("message", "Error fetching run from state, acking", "run_id", runReceipt.Run.RunID, "error", fmt.Sprintf("%+v", err))
			if err = runReceipt.Done(); err != nil {
				sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
			}
			continue
		}

		//
		// Fetch run's definition from state manager
		//
		// * Will not be necessary once we copy relevant run information from definition onto the run itself
		//
		var executable state.Executable
		executableType := run.ExecutableType

		// If the executable type is not set, safe to assume that it is a legacy
		// definition.
		if executableType == nil {
			executable, err = sw.sm.GetDefinition(run.DefinitionID)
		} else {
			switch *executableType {
			case state.ExecutableTypeDefinition:
			default:
				executable, err = sw.sm.GetDefinition(run.DefinitionID)
			}
		}

		if err != nil {
			sw.log.Log(
				"message", "Error fetching definition for run",
				"run_id", run.RunID,
				"definition_id", run.DefinitionID,
				"error", err.Error())
			if err = runReceipt.Done(); err != nil {
				sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
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
			launched, retryable, err := sw.ee.Execute(executable, run, sw.sm)
			if err != nil {
				sw.log.Log("message", "Error executing run", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err), "retryable", retryable)
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
			err = sw.log.Event("eventClassName", "FlotillaSubmitTask", "definition", executable, "run_id", run.RunID)
			if err != nil {
				sw.log.Log("message", "Failed to emit event", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
			}

			//
			// UpdateStatus the status and information of the run;
			// either the run submitted successfully -or- it did not and is not retryable
			//
			if _, err = sw.sm.UpdateRun(run.RunID, launched); err != nil {
				sw.log.Log("message", "Failed to update run status", "run_id", run.RunID, "status", launched.Status, "error", fmt.Sprintf("%+v", err))
			}
		} else {
			sw.log.Log("message", "Received run that is not runnable", "run_id", run.RunID, "status", run.Status)
		}

		if err = runReceipt.Done(); err != nil {
			sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
		}
	}
}
