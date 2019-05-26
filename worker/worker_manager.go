package worker

import (
	"github.com/stitchfix/flotilla-os/state"
)

type workerManager struct {
	Worker
	sm state.Manager
}

func (wm *workerManager) Initialize(sm state.Manager) {
	wm.sm = sm
}

func (wm *workerManager) getWorkersList() (state.WorkersList, error) {
	return wm.sm.ListWorkers()
}
