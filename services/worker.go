package services

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

//
// WorkerService defines an interface for operations involving workers
//
type WorkerService interface {
	List() (state.WorkersList, error)
}

type workerService struct {
	sm state.Manager
}

//
// NewWorkerService configures and returns a WorkerService
//
func NewWorkerService(conf config.Config, sm state.Manager) (WorkerService, error) {
	ws := workerService{sm: sm}
	return &ws, nil
}

func (ws *workerService) List() (state.WorkersList, error) {
	return ws.sm.ListWorkers()
}
