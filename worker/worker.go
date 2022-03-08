package worker

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/queue"
	"time"

	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
)

//
// Worker defines a background worker process
//
type Worker interface {
	Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager) error
	Run() error
	GetTomb() *tomb.Tomb
}

//
// NewWorker instantiates a new worker.
//
func NewWorker(workerType string, log flotillaLog.Logger, conf config.Config, eksEngine engine.Engine, emrEngine engine.Engine, sm state.Manager, qm queue.Manager) (Worker, error) {
	var worker Worker

	switch workerType {
	case "submit":
		worker = &submitWorker{}
	case "retry":
		worker = &retryWorker{}
	case "status":
		worker = &statusWorker{}
	case "worker_manager":
		worker = &workerManager{}
	case "cloudtrail":
		worker = &cloudtrailWorker{}
	case "events":
		worker = &eventsWorker{}
	default:
		return nil, errors.Errorf("no workerType [%s] exists", workerType)
	}

	pollInterval, err := GetPollInterval(workerType, conf)
	if err = worker.Initialize(conf, sm, eksEngine, emrEngine, log, pollInterval, qm); err != nil {
		return worker, errors.Wrapf(err, "problem initializing worker [%s]", workerType)
	}
	return worker, nil
}

//
// GetPollInterval returns the frequency at which a worker will run.
//
func GetPollInterval(workerType string, conf config.Config) (time.Duration, error) {
	var interval time.Duration
	pollIntervalString := conf.GetString(fmt.Sprintf("worker_%s_interval", workerType))
	if len(pollIntervalString) == 0 {
		return interval, errors.Errorf("worker type: [%s] needs worker_%s_interval set", workerType, workerType)
	}
	return time.ParseDuration(pollIntervalString)
}
