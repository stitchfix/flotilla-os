package worker

import (
	"context"
	"fmt"
	"github.com/stitchfix/flotilla-os/tracing"

	"github.com/stitchfix/flotilla-os/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"time"

	"github.com/go-redis/redis"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
)

type submitWorker struct {
	sm             state.Manager
	eksEngine      engine.Engine
	emrEngine      engine.Engine
	conf           config.Config
	log            flotillaLog.Logger
	pollInterval   time.Duration
	t              tomb.Tomb
	redisClient    *redis.Client
	clusterManager *engine.DynamicClusterManager
}

func (sw *submitWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager, clusterManager *engine.DynamicClusterManager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.eksEngine = eksEngine
	sw.emrEngine = emrEngine
	sw.log = log
	sw.redisClient, _ = utils.SetupRedisClient(conf)
	sw.clusterManager = clusterManager
	_ = sw.log.Log("level", "info", "message", "initialized a submit worker")
	return nil
}

func (sw *submitWorker) GetTomb() *tomb.Tomb {
	return &sw.t
}

// Run lists queues, consumes runs from them, and executes them using the execution engine
func (sw *submitWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-sw.t.Dying():
			sw.log.Log("level", "info", "message", "A submit worker was terminated")
			return nil
		default:
			sw.runOnce(ctx)
			time.Sleep(sw.pollInterval)
		}
	}
}
func (sw *submitWorker) runOnce(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "submit_worker.poll", "submit_worker")
	defer span.Finish()
	var receipts []engine.RunReceipt
	var run state.Run
	var err error

	pollStart := time.Now()
	receipts, err = sw.eksEngine.PollRuns(ctx)
	span.SetTag("sqs.poll_duration_ms", time.Since(pollStart).Milliseconds())
	span.SetTag("sqs.received_count", len(receipts))
	receiptsEMR, err := sw.emrEngine.PollRuns(ctx)
	receipts = append(receipts, receiptsEMR...)
	if err != nil {
		sw.log.Log("level", "error", "message", "Error receiving runs", "error", fmt.Sprintf("%+v", err))
	}
	for _, runReceipt := range receipts {
		if runReceipt.Run == nil {
			continue
		}
		sw.log.Log("level", "info", "message", "Processing run receipt",
			"run_id", runReceipt.Run.RunID,
			"has_trace_context", runReceipt.TraceID != 0 && runReceipt.ParentID != 0,
			"trace_id", runReceipt.TraceID,
			"parent_id", runReceipt.ParentID)

		var runCtx context.Context
		if runReceipt.RunReceipt.TraceID != 0 && runReceipt.RunReceipt.ParentID != 0 {
			carrier := tracing.TextMapCarrier{
				"x-datadog-trace-id":          fmt.Sprintf("%d", runReceipt.TraceID),
				"x-datadog-parent-id":         fmt.Sprintf("%d", runReceipt.ParentID),
				"x-datadog-sampling-priority": fmt.Sprintf("%d", runReceipt.SamplingPriority),
			}
			spanCtx, err := tracer.Extract(carrier)
			if err != nil {
				sw.log.Log("level", "error", "message", "Error extracting span context", "error", err.Error())
				runCtx = ctx
			} else {
				bridgeSpan := tracer.StartSpan("flotilla.queue.sqs_receive", tracer.ChildOf(spanCtx))
				bridgeSpan.SetTag("run_id", runReceipt.Run.RunID)
				runCtx = tracer.ContextWithSpan(ctx, bridgeSpan)
				defer bridgeSpan.Finish()
			}
		} else {
			runCtx = ctx
		}
		runCtx, childSpan := utils.TraceJob(runCtx, "flotilla.job.submit_worker.process", "")
		childSpan.SetTag("job.run_id", runReceipt.Run.RunID)
		utils.TagJobRun(childSpan, *runReceipt.Run)

		//
		// Fetch run from state manager to ensure its existence
		//
		run, err = sw.sm.GetRun(ctx, runReceipt.Run.RunID)
		if err != nil {
			sw.log.Log("level", "error", "message", "Error fetching run from state, acking", "run_id", runReceipt.Run.RunID, "error", fmt.Sprintf("%+v", err))
			if err = runReceipt.Done(); err != nil {
				sw.log.Log("level", "error", "message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
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
				d, err = sw.sm.GetDefinition(runCtx, *run.ExecutableID)

				if err != nil {
					sw.logFailedToGetExecutableMessage(run, err)
					if err = runReceipt.Done(); err != nil {
						sw.log.Log("level", "error", "message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
					}
					continue
				}

				// Execute the run using the execution engine.
				if run.Engine == nil || *run.Engine == state.EKSEngine {
					launched, retryable, err = sw.eksEngine.Execute(runCtx, d, run, sw.sm)
				} else {
					launched, retryable, err = sw.emrEngine.Execute(runCtx, d, run, sw.sm)
				}

				break
			case state.ExecutableTypeTemplate:
				var tpl state.Template
				tpl, err = sw.sm.GetTemplateByID(runCtx, *run.ExecutableID)

				if err != nil {
					sw.logFailedToGetExecutableMessage(run, err)
					if err = runReceipt.Done(); err != nil {
						sw.log.Log("level", "error", "message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
					}
					continue
				}

				// Execute the run using the execution engine.
				sw.log.Log("level", "info", "message", "Submitting", "run_id", run.RunID)
				launched, retryable, err = sw.eksEngine.Execute(runCtx, tpl, run, sw.sm)
				break
			default:
				// If executable type is invalid; log message and continue processing
				// other runs.
				sw.log.Log("level", "error", "message", "submit worker failed", "run_id", run.RunID, "error", "invalid executable type")
				continue
			}

			if err != nil {
				sw.log.Log("level", "error", "message", "Error executing run", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err), "retryable", retryable)
				if !retryable {
					// Set status to StatusStopped, and ack
					launched.Status = state.StatusStopped
				} else {
					// Don't change status, don't ack
					continue
				}
			} else {
				sw.log.Log("level", "info", "message", "Task submitted from SQS to the cluster", "run_id", run.RunID)
			}

			//
			// Emit event with current definition
			//
			err = sw.log.Event("eventClassName", "FlotillaSubmitTask", "executable_id", *run.ExecutableID, "run_id", run.RunID)
			if err != nil {
				sw.log.Log("level", "error", "message", "Failed to emit event", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
			}

			//
			// UpdateStatus the status and information of the run;
			// either the run submitted successfully -or- it did not and is not retryable
			//
			if _, err = sw.sm.UpdateRun(runCtx, run.RunID, launched); err != nil {
				sw.log.Log("level", "error", "message", "Failed to update run status", "run_id", run.RunID, "status", launched.Status, "error", fmt.Sprintf("%+v", err))
			}
		} else {
			sw.log.Log("level", "warn", "message", "Received run that is not runnable", "run_id", run.RunID, "status", run.Status)
		}

		if err = runReceipt.Done(); err != nil {
			childSpan.SetTag("error", true)
			childSpan.SetTag("error.msg", err.Error())
			childSpan.SetTag("error.type", "sqs_ack")
			sw.log.Log("level", "error", "message", "Acking run failed", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
		} else {
			childSpan.SetTag("sqs.ack_success", true)
		}
		childSpan.Finish()
	}
}

func (sw *submitWorker) logFailedToGetExecutableMessage(run state.Run, err error) {
	sw.log.Log(
		"level", "error",
		"message", "Error fetching executable for run",
		"run_id", run.RunID,
		"executable_id", run.ExecutableID,
		"executable_type", run.ExecutableType,
		"error", err.Error())
}
