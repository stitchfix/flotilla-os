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
	submitWorkers []Worker
	retryWorkers []Worker
	statusWorkers []Worker
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

			// Append worker to the appropriate worker slice.
			switch w.WorkerType {
			case "retry":
				wm.retryWorkers = append(wm.retryWorkers, wk)
			case "status":
				wm.statusWorkers = append(wm.statusWorkers, wk)
			case "submit":
				wm.submitWorkers = append(wm.submitWorkers, wk)
			default:
				// todo handle error
				fmt.Println("invalid worker type")
			}
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
	rwLen := len(wm.retryWorkers)
	sbLen := len(wm.submitWorkers)
	stLen := len(wm.statusWorkers)

	if err != nil {
		return err
	}

	for _, w := range workerList.Workers {
		switch w.WorkerType {
		case "retry":
			if w.CountPerInstance != rwLen {
				if err := wm.updateWorkerCount("retry", rwLen, w.CountPerInstance); err != nil {
					// log error
				}
			}
		case "status":
			if w.CountPerInstance != stLen {
				if err := wm.updateWorkerCount("status", stLen, w.CountPerInstance); err != nil {
					// log error
				}
			}
		case "submit":
			if w.CountPerInstance != sbLen {
				if err := wm.updateWorkerCount("submit", sbLen, w.CountPerInstance); err != nil {
					// log error
				}
			}
		}
	}

	return nil
}

func (wm *workerManager) updateWorkerCount(workerType string, curr int, next int) error {
	var wSlice *[]Worker

	switch workerType {
	case "retry":
		wSlice = &wm.retryWorkers
	case "status":
		wSlice = &wm.statusWorkers
	case "submit":
		wSlice = &wm.submitWorkers
	default:
		return nil
	}

	if curr > next {
		// Kill workers
		for i := next; i < curr-1; i++ {
			(*wSlice)[i].GetTomb().Kill(nil)
		}
	} else if curr < next {
		// Add workers
		for i := curr; i < next; i++ {
			wk, err := NewWorker(workerType, wm.log, wm.conf, wm.ee, wm.sm)

			if err != nil {
				return err
			}

			// Start goroutine via tomb
			wk.GetTomb().Go(wk.Run)

			// Append it
			*wSlice = append(*wSlice, wk)
		}
	}
	return nil
}
