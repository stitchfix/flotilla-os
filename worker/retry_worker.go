package worker

import (
	"context"
	"fmt"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/utils"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
)

type retryWorker struct {
	sm             state.Manager
	ee             engine.Engine
	conf           config.Config
	log            flotillaLog.Logger
	pollInterval   time.Duration
	t              tomb.Tomb
	clusterManager *engine.DynamicClusterManager
}

func (rw *retryWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager, clusterManager *engine.DynamicClusterManager) error {
	rw.pollInterval = pollInterval
	rw.conf = conf
	rw.sm = sm
	rw.ee = eksEngine
	rw.log = log
	rw.clusterManager = clusterManager
	rw.log.Log("message", "initialized a retry worker")
	return nil
}

func (rw *retryWorker) GetTomb() *tomb.Tomb {
	return &rw.t
}

// Run finds tasks that NEED_RETRY and requeues them
func (rw *retryWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-rw.t.Dying():
			rw.log.Log("message", "A retry worker was terminated")
			return nil
		default:
			rw.runOnce(ctx)
			time.Sleep(rw.pollInterval)
		}
	}
}

func (rw *retryWorker) runOnce(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "flotilla.retry_worker.poll", "retry_worker")
	defer span.Finish()
	// List runs in the StatusNeedsRetry state and requeue them
	runList, err := rw.sm.ListRuns(ctx, 25, 0, "started_at", "asc", map[string][]string{"status": {state.StatusNeedsRetry}}, nil, []string{state.EKSEngine})
	if runList.Total > 0 {
		rw.log.Log("message", fmt.Sprintf("Got %v jobs to retry", runList.Total))
	}

	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		rw.log.Log("message", "Error listing runs for retry", "error", fmt.Sprintf("%+v", err))
		return
	}

	for _, run := range runList.Runs {
		_, childSpan := utils.TraceJob(ctx, "flotilla.job.retry", run.RunID)
		func() {
			defer childSpan.Finish()
			utils.TagJobRun(childSpan, run)

			if _, err = rw.sm.UpdateRun(ctx, run.RunID, state.Run{Status: state.StatusQueued}); err != nil {
				rw.log.Log("message", "Error updating run status to StatusQueued", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
				return
			}

			if err = rw.ee.Enqueue(ctx, run); err != nil {
				rw.log.Log("message", "Error enqueuing run", "run_id", run.RunID, "error", fmt.Sprintf("%+v", err))
				return
			}
		}()
	}
	return
}
