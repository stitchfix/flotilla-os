package worker

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/tomb.v2"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
)

type workerManager struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	workers map[string][]Worker
	t tomb.Tomb
}

func (wm *workerManager) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration) error {
	wm.conf = conf
	wm.log = log
	wm.ee = ee
	wm.sm = sm
	wm.pollInterval = pollInterval

	if err := wm.InitializeWorkers(); err != nil {
		return errors.Errorf("WorkerManager unable to initialize workers.")
	}

	return nil
}

func (wm *workerManager) GetTomb() *tomb.Tomb {
	return &wm.t
}

//
// InitializeWorkers will first check the DB for the total count per instance
// of each worker type (retry, submit, or status), start each worker's  `Run`
// goroutine via tomb, then append the worker to the appropriate slice.
//
func (wm *workerManager) InitializeWorkers() error {
	workerList, err := wm.sm.ListWorkers()

	if err != nil {
		return err
	}
	wm.workers = make(map[string][]Worker)

	// Iterate through list of workers.
	for _, w := range workerList.Workers {
		for i := 0; i < w.CountPerInstance; i++ {
			// Instantiate a new worker.
			wk, err := NewWorker(w.WorkerType, wm.log, wm.conf, wm.ee, wm.sm)

			if err != nil {
				return err
			}

			// Start goroutine via tomb
			wk.GetTomb().Go(wk.Run)
			wm.workers[w.WorkerType] = append(wm.workers[w.WorkerType], wk)
		}
	}

	return nil
}

func (wm *workerManager) Run() error {
	for {
		select {
		case <-wm.t.Dying():
			wm.log.Log("message", "Worker manager was terminated")
			return nil
		default:
			wm.runOnce()
			time.Sleep(wm.pollInterval)
		}
	}
}

func (wm *workerManager) runOnce() error {
	// Check worker count via state manager.
	workerList, err := wm.sm.ListWorkers()

	if err != nil {
		return err
	}

	for _, w := range workerList.Workers {
		if len(wm.workers[w.WorkerType]) != w.CountPerInstance {
			if err := wm.updateWorkerCount(w.WorkerType, len(wm.workers[w.WorkerType]), w.CountPerInstance); err != nil {
				// log error
			}
		}
	}

	return nil
}

func (wm *workerManager) updateWorkerCount(workerType string, curr int, next int) error {
	if curr > next {
		// Kill workers
		for i := next; i < curr; i++ {
			if err := wm.removeWorker(workerType); err != nil {
				return err
			}
		}
	} else if curr < next {
		// Add workers
		for i := curr; i < next; i++ {
			if err := wm.addWorker(workerType); err != nil {
				return err
			}
		}
	}
	return nil
}

func (wm *workerManager) removeWorker(workerType string) error {
	if workers, ok := wm.workers[workerType]; ok {
		if len(workers) > 0 {
			toKill := workers[len(workers)-1]
			toKill.GetTomb().Kill(nil)
			wm.workers[workerType] = workers[:len(workers)-1]
		}
	} else {
		return fmt.Errorf("invalid worker type %s", workerType)
	}
	return nil
}

func (wm *workerManager) addWorker(workerType string) error {
	wk, err := NewWorker(workerType, wm.log, wm.conf, wm.ee, wm.sm)

	if err != nil {
		return err
	}

	// Start goroutine via tomb
	wk.GetTomb().Go(wk.Run)
	if workers, ok := wm.workers[workerType]; ok {
		workers = append(workers, wk)
	} else {
		return fmt.Errorf("invalid worker type %s", workerType)
	}
	return nil
}