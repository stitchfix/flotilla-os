package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/utils"
	"gopkg.in/tomb.v2"
)

type retentionWorker struct {
	sm           state.Manager
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
}

func (rw *retentionWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager, clusterManager *engine.DynamicClusterManager) error {
	rw.pollInterval = pollInterval
	rw.conf = conf
	rw.sm = sm
	rw.log = log
	rw.log.Log("level", "info", "message", "initialized a retention worker")
	return nil
}

func (rw *retentionWorker) GetTomb() *tomb.Tomb {
	return &rw.t
}

func (rw *retentionWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-rw.t.Dying():
			rw.log.Log("level", "info", "message", "A retention worker was terminated")
			return nil
		default:
			rw.runOnce(ctx)
			time.Sleep(rw.pollInterval)
		}
	}
}

func (rw *retentionWorker) runOnce(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "flotilla.retention_worker.poll", "retention_worker")
	defer span.Finish()

	retentionDays := 90
	if rw.conf.IsSet("task_retention_days") {
		retentionDays = rw.conf.GetInt("task_retention_days")
	}

	deleted, err := rw.sm.DeleteOldRuns(ctx, retentionDays)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		rw.log.Log("level", "error", "message", "Error deleting old runs", "error", fmt.Sprintf("%+v", err))
		return
	}

	if deleted > 0 {
		rw.log.Log("level", "info", "message", fmt.Sprintf("Deleted %d old task rows (retention: %d days)", deleted, retentionDays))
	}
}
