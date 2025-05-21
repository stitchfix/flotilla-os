package worker

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/utils"
	"gopkg.in/tomb.v2"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
)

type workerManager struct {
	sm             state.Manager
	eksEngine      engine.Engine
	emrEngine      engine.Engine
	conf           config.Config
	log            flotillaLog.Logger
	pollInterval   time.Duration
	workers        map[string][]Worker
	t              tomb.Tomb
	engine         *string
	qm             queue.Manager
	clusterManager *engine.DynamicClusterManager
}

func (wm *workerManager) Initialize(
	conf config.Config,
	sm state.Manager,
	eksEngine engine.Engine,
	emrEngine engine.Engine,
	log flotillaLog.Logger,
	pollInterval time.Duration,
	qm queue.Manager,
	clusterManager *engine.DynamicClusterManager,
) error {
	wm.conf = conf
	wm.log = log
	wm.eksEngine = eksEngine
	wm.emrEngine = emrEngine
	wm.sm = sm
	wm.qm = qm
	wm.pollInterval = pollInterval
	wm.clusterManager = clusterManager

	ctx, span := utils.TraceJob(context.Background(), "worker_manager.initialize_workers", "worker_manager")
	defer span.Finish()

	if err := wm.InitializeWorkers(ctx); err != nil {
		span.SetTag("error", err.Error())
		return errors.Errorf("WorkerManager unable to initialize workers: %s", err.Error())
	}
	return nil
}

func (wm *workerManager) GetTomb() *tomb.Tomb {
	return &wm.t
}

// InitializeWorkers will first check the DB for the total count per instance
// of each worker type (retry, submit, or status), start each worker's  `Run`
// goroutine via tomb, then append the worker to the appropriate slice.
func (wm *workerManager) InitializeWorkers(ctx context.Context) error {
	workerList, err := wm.sm.ListWorkers(ctx, state.EKSEngine)

	if err != nil {
		return err
	}

	wm.workers = make(map[string][]Worker)

	// Iterate through list of workers.
	for _, w := range workerList.Workers {
		wm.workers[w.WorkerType] = make([]Worker, w.CountPerInstance)
		for i := 0; i < w.CountPerInstance; i++ {
			// Instantiate a new worker.
			wk, err := NewWorker(w.WorkerType, wm.log, wm.conf, wm.eksEngine, wm.emrEngine, wm.sm, wm.qm, wm.clusterManager)

			if err != nil {
				return err
			}

			// Start goroutine via tomb
			wk.GetTomb().Go(func() error {
				return wk.Run(ctx)
			})
			wm.workers[w.WorkerType][i] = wk
		}
	}

	return nil
}

func (wm *workerManager) Run(ctx context.Context) error {
	for {
		select {
		case <-wm.t.Dying():
			wm.log.Log("message", "Worker manager was terminated")
			return nil
		default:
			ctx, span := utils.TraceJob(context.Background(), "worker_manager.run_once", "worker_manager")
			wm.runOnce(ctx)
			span.Finish()
			time.Sleep(wm.pollInterval)
		}
	}
}

func (wm *workerManager) runOnce(ctx context.Context) error {
	// Check worker count via state manager.
	workerList, err := wm.sm.ListWorkers(ctx, state.EKSEngine)

	if err != nil {
		return err
	}

	for _, w := range workerList.Workers {
		currentWorkerCount := len(wm.workers[w.WorkerType])
		// Is our current number of workers not the desired number of workers?
		if currentWorkerCount != w.CountPerInstance {

			if err := wm.updateWorkerCount(ctx, w.WorkerType, currentWorkerCount, w.CountPerInstance); err != nil {
				wm.log.Log(
					"message", "problem updating worker count",
					"error", err.Error())
			}
		}
	}

	return nil
}

func (wm *workerManager) updateWorkerCount(
	ctx context.Context,
	workerType string,
	currentWorkerCount int,
	desiredWorkerCount int,
) error {
	ctx, span := utils.TraceJob(ctx, "worker_manager.update_worker_count", workerType)
	defer span.Finish()

	//span.SetTag("current_count", currentWorkerCount)
	//span.SetTag("desired_count", desiredWorkerCount)

	if currentWorkerCount > desiredWorkerCount {
		for i := desiredWorkerCount; i < currentWorkerCount; i++ {
			wm.log.Log("message", fmt.Sprintf(
				"Scaling down %s workers from %d to %d", workerType, currentWorkerCount, desiredWorkerCount))
			if err := wm.removeWorker(ctx, workerType); err != nil {
				return err
			}
		}
	} else if currentWorkerCount < desiredWorkerCount {
		for i := currentWorkerCount; i < desiredWorkerCount; i++ {
			wm.log.Log("message", fmt.Sprintf(
				"Scaling up %s workers from %d to %d", workerType, currentWorkerCount, desiredWorkerCount))
			if err := wm.addWorker(ctx, workerType); err != nil {
				return err
			}
		}
	}
	return nil
}

func (wm *workerManager) removeWorker(ctx context.Context, workerType string) error {
	ctx, span := utils.TraceJob(ctx, "worker_manager.remove_worker", workerType)
	defer span.Finish()

	if workers, ok := wm.workers[workerType]; ok {
		if len(workers) > 0 {
			toKill := workers[len(workers)-1]
			toKill.GetTomb().Kill(nil)
			wm.workers[workerType] = workers[:len(workers)-1]
			wm.log.Log("message", "Removed worker", "type", workerType)
		}
	} else {
		return fmt.Errorf("invalid worker type %s", workerType)
	}
	return nil
}

func (wm *workerManager) addWorker(ctx context.Context, workerType string) error {
	ctx, span := utils.TraceJob(ctx, "worker_manager.add_worker", workerType)
	defer span.Finish()

	wk, err := NewWorker(workerType, wm.log, wm.conf, wm.eksEngine, wm.emrEngine, wm.sm, wm.qm, wm.clusterManager)
	if err != nil {
		return err
	}
	wk.GetTomb().Go(func() error {
		return wk.Run(ctx)
	})
	if _, ok := wm.workers[workerType]; ok {
		wm.workers[workerType] = append(wm.workers[workerType], wk)
	} else {
		return fmt.Errorf("invalid worker type %s", workerType)
	}
	wm.log.Log("message", "Added worker", "type", workerType)
	return nil
}
