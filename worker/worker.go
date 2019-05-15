package worker

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
)

//
// WorkerTypes is an enum for the different types of workers.
//
var WorkerTypes = newWorkerTypes()

type workerTypes struct {
	Retry  string
	Submit string
	Status string
}

func newWorkerTypes() *workerTypes {
	return &workerTypes{
		Retry:  "retry",
		Submit: "submit",
		Status: "status",
	}
}

//
// Worker defines a background worker process
//
type Worker interface {
	Initialize(
		conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration) error
	Run()
}

//
// NewWorker instantiates a new worker.
//
func NewWorker(
	workerType string,
	log flotillaLog.Logger,
	conf config.Config,
	ee engine.Engine,
	sm state.Manager) (Worker, error) {

	var worker Worker

	switch workerType {
	case WorkerTypes.Submit:
		worker = &submitWorker{}
	case WorkerTypes.Retry:
		worker = &retryWorker{}
	case WorkerTypes.Status:
		worker = &statusWorker{}
	default:
		return nil, errors.Errorf("no workerType [%s] exists", workerType)
	}

	pollInterval, err := GetPollInterval(workerType, conf)
	if err = worker.Initialize(conf, sm, ee, log, pollInterval); err != nil {
		return worker, errors.Wrapf(err, "problem initializing worker [%s]", workerType)
	}
	return worker, nil
}

//
// GetPollInterval returns the frequency at which a worker will run.
//
func GetPollInterval(workerType string, conf config.Config) (time.Duration, error) {
	var interval time.Duration
	pollIntervalString := conf.GetString(fmt.Sprintf("worker.%s_interval", workerType))
	if len(pollIntervalString) == 0 {
		return interval, errors.Errorf("worker type: [%s] needs worker.%s_interval set", workerType, workerType)
	}
	return time.ParseDuration(pollIntervalString)
}
