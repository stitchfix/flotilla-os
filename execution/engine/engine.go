package engine

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Engine defines the execution engine interface.
//
type Engine interface {
	Initialize(conf config.Config) error
	Execute(executable state.Executable, run state.Run, manager state.Manager) (state.Run, bool, error)
	Terminate(run state.Run) error
	Enqueue(run state.Run) error
	PollRuns() ([]RunReceipt, error)
	PollRunStatus() (state.Run, error)
	PollStatus() (RunReceipt, error)
	GetEvents(run state.Run) (state.PodEventList, error)
	FetchUpdateStatus(run state.Run) (state.Run, error)
	FetchPodMetrics(run state.Run) (state.Run, error)

	// Legacy methods from the ECS era. Here for backwards compatibility.
	Define(definition state.Definition) (state.Definition, error)
	Deregister(definition state.Definition) error
}

type RunReceipt struct {
	queue.RunReceipt
}

//
// NewExecutionEngine initializes and returns a new Engine
//
func NewExecutionEngine(conf config.Config, qm queue.Manager, name string, logger log.Logger) (Engine, error) {
	switch name {
	case "k8s":
		k8sEng := &K8SExecutionEngine{qm: qm, log: logger}
		if err := k8sEng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing ECSExecutionEngine")
		}
		return k8sEng, nil
	default:
		return nil, fmt.Errorf("no Engine named [%s] was found", name)
	}
}
