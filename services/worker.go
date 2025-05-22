package services

import (
	"context"
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

//
// WorkerService defines an interface for operations involving workers
//
type WorkerService interface {
	List(ctx context.Context, engine string) (state.WorkersList, error)
	Get(ctx context.Context, workerType string, engine string) (state.Worker, error)
	Update(ctx context.Context, workerType string, updates state.Worker) (state.Worker, error)
	BatchUpdate(ctx context.Context, updates []state.Worker) (state.WorkersList, error)
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

func (ws *workerService) List(ctx context.Context, engine string) (state.WorkersList, error) {
	return ws.sm.ListWorkers(ctx, engine)
}

func (ws *workerService) Get(ctx context.Context, workerType string, engine string) (state.Worker, error) {
	var w state.Worker
	if err := ws.validate(workerType); err != nil {
		return w, err
	}
	return ws.sm.GetWorker(ctx, workerType, engine)
}

func (ws *workerService) Update(ctx context.Context, workerType string, updates state.Worker) (state.Worker, error) {
	var w state.Worker
	if err := ws.validate(workerType); err != nil {
		return w, err
	}

	return ws.sm.UpdateWorker(ctx, workerType, updates)
}

func (ws *workerService) BatchUpdate(ctx context.Context, updates []state.Worker) (state.WorkersList, error) {
	var wl state.WorkersList
	for _, update := range updates {
		if err := ws.validate(update.WorkerType); err != nil {
			return wl, err
		}
	}
	return ws.sm.BatchUpdateWorkers(ctx, updates)
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
