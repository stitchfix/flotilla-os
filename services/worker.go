package services

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

//
// WorkerService defines an interface for operations involving workers
//
type WorkerService interface {
	List() (state.WorkersList, error)
	Get(workerType string) (state.Worker, error)
	Update(workerType string, updates state.Worker) (state.Worker, error)
	BatchUpdate(updates []state.Worker) (state.WorkersList, error)
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

func (ws *workerService) Get(workerType string) (state.Worker, error) {
	var w state.Worker
	if err := ws.validate(workerType); err != nil {
		return w, err
	}
	return ws.sm.GetWorker(workerType)
}

func (ws *workerService) Update(workerType string, updates state.Worker) (state.Worker, error) {
	var w state.Worker
	if err := ws.validate(workerType); err != nil {
		return w, err
	}

	return ws.sm.UpdateWorker(workerType, updates)
}

func (ws *workerService) BatchUpdate(updates []state.Worker) (state.WorkersList, error) {
	var wl state.WorkersList
	for _, update := range updates {
		if err := ws.validate(update.WorkerType); err != nil {
			return wl, err
		}
	}
	return ws.sm.BatchUpdateWorkers(updates)
}

func (ws *workerService) validate(workerType string) error {
	if !state.IsValidWorkerType(workerType) {
		var validTypesList []string
		for validType := range state.WorkerTypes {
			validTypesList = append(validTypesList, validType)
		}
		return exceptions.MalformedInput{
			ErrorString: fmt.Sprintf(
				"Worker type: [%s] is not a valid worker type; valid types: %s",
				workerType, validTypesList)}
	}
	return nil
}
