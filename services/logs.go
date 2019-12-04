package services

import (
	"github.com/stitchfix/flotilla-os/clients/logs"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"net/http"
)

type LogService interface {
	Logs(runID string, lastSeen *string) (string, *string, error)
	LogsText(runID string, w http.ResponseWriter) error
}

type logService struct {
	sm state.Manager
	lc logs.Client
}

func NewLogService(conf config.Config, sm state.Manager, lc logs.Client) (LogService, error) {
	return &logService{sm: sm, lc: lc}, nil
}

func (ls *logService) Logs(runID string, lastSeen *string) (string, *string, error) {
	run, err := ls.sm.GetRun(runID)
	if err != nil {
		return "", nil, err
	}

	if run.Status != state.StatusRunning && run.Status != state.StatusStopped {
		// Won't have logs yet
		return "", nil, nil
	}

	defn, err := ls.sm.GetDefinition(run.DefinitionID)
	if err != nil {
		return "", nil, err
	}

	return ls.lc.Logs(defn, run, lastSeen)
}

func (ls *logService) LogsText(runID string, w http.ResponseWriter) error {
	run, err := ls.sm.GetRun(runID)
	if err != nil {
		return err
	}

	if run.Status != state.StatusRunning && run.Status != state.StatusStopped {
		// Won't have logs yet
		return nil
	}

	defn, err := ls.sm.GetDefinition(run.DefinitionID)
	if err != nil {
		return err
	}

	return ls.lc.LogsText(defn, run, w)
}
