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
	eksEngine    engine.Engine
	emrEngine    engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	redisClient  *redis.Client
}

func (sw *submitWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.eksEngine = eksEngine
	sw.emrEngine = emrEngine
	sw.log = log
	sw.redisClient = redis.NewClient(&redis.Options{Addr: conf.GetString("redis_address"), DB: conf.GetInt("redis_db")})
	_ = sw.log.Log("message", "initialized a submit worker")
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
	var receipts []engine.RunReceipt
	var run state.Run
	var err error

	receipts, err = sw.eksEngine.PollRuns()
	receiptsEMR, err := sw.emrEngine.PollRuns()
	receipts = append(receipts, receiptsEMR...)
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
		run, err = sw.sm.GetRun(runReceipt.Run.RunID)
		if err != nil {
			sw.log.Log("message", "Error fetching run from state, acking", "run_id", runReceipt.Run.RunID, "error", fmt.Sprintf("%+v", err))
			if err = runReceipt.Done(); err != nil {
				sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
			}
			continue
		}

		//
		// Only valid to process if it's in the StatusQueued state
		//
		if run.Status == state.StatusQueued {
			var (
				launched  state.Run
				retryable bool
			)

			// 1. Check for existence of run.ExecutableType; set to `task_definition`
			// if not set.
			if run.ExecutableType == nil {
				defaultExecutableType := state.ExecutableTypeDefinition
				run.ExecutableType = &defaultExecutableType
			}

			// 2. Check for existence of run.ExecutableID; set to run.DefinitionID if
			// not set.
			if run.ExecutableID == nil {
				defID := run.DefinitionID
				run.ExecutableID = &defID
			}

			// 3. Switch by executable type.
			switch *run.ExecutableType {
			case state.ExecutableTypeDefinition:
				var d state.Definition
				d, err = sw.sm.GetDefinition(*run.ExecutableID)

				if err != nil {
					sw.logFailedToGetExecutableMessage(run, err)
					if err = runReceipt.Done(); err != nil {
						sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
					}
					continue
				}

				// Execute the run using the execution engine.
				if run.Engine == nil || *run.Engine == state.EKSEngine {
					launched, retryable, err = sw.eksEngine.Execute(d, run, sw.sm)
				} else {
					launched, retryable, err = sw.emrEngine.Execute(d, run, sw.sm)
				}

				break
			case state.ExecutableTypeTemplate:
				var tpl state.Template
				tpl, err = sw.sm.GetTemplateByID(*run.ExecutableID)

				if err != nil {
					sw.logFailedToGetExecutableMessage(run, err)
					if err = runReceipt.Done(); err != nil {
						sw.log.Log("message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
					}
					continue
				}

				// Execute the run using the execution engine.
				sw.log.Log("message", "Submitting", "run_id", run.RunID)
				launched, retryable, err = sw.eksEngine.Execute(tpl, run, sw.sm)
				break
			default:
				// If executable type is invalid; log message and continue processing
				// other runs.
				sw.log.Log("message", "submit worker failed", "run_id", run.RunID, "error", "invalid executable type")
				continue
			}

			if err != nil {
				sw.log.Log("message", "Error executing run", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err), "retryable", retryable)
				if !retryable {
					// Set status to StatusStopped, and ack
					launched.Status = state.StatusStopped
				} else {
					// Don't change status, don't ack
					continue
				}
			} else {
				sw.log.Log("message", "Task submitted from SQS to the cluster", "run_id", run.RunID)
			}

			//
			// Emit event with current definition
			//
			err = sw.log.Event("eventClassName", "FlotillaSubmitTask", "executable_id", *run.ExecutableID, "run_id", run.RunID)
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

func (sw *submitWorker) logFailedToGetExecutableMessage(run state.Run, err error) {
	sw.log.Log(
		"message", "Error fetching executable for run",
		"run_id", run.RunID,
		"executable_id", run.ExecutableID,
		"executable_type", run.ExecutableType,
		"error", err.Error())
}
