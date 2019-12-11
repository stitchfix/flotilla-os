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
	// v0
	Execute(definition state.Definition, run state.Run, manager state.Manager) (state.Run, bool, error)

	// v1 - once runs contain a copy of relevant definition info
	// Execute(run state.Run) error

	Define(definition state.Definition) (state.Definition, error)

	Deregister(definition state.Definition) error

	Terminate(run state.Run) error

	Enqueue(run state.Run) error

	PollRuns() ([]RunReceipt, error)

	PollStatus() (RunReceipt, error)

	GetEvents(run state.Run) (state.PodEventList, error)

	FetchUpdateStatus(run state.Run)(state.Run, error)
}

type RunReceipt struct {
	queue.RunReceipt
}

//
// NewExecutionEngine initializes and returns a new Engine
//
func NewExecutionEngine(conf config.Config, qm queue.Manager, name string, logger log.Logger) (Engine, error) {
	switch name {
	case "ecs":
		eng := &ECSExecutionEngine{qm: qm}
		if err := eng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing ECSExecutionEngine")
		}
		return eng, nil
	case "eks":
		eksEng := &EKSExecutionEngine{qm: qm, log: logger}
		if err := eksEng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing ECSExecutionEngine")
		}
		return eksEng, nil
	default:
		return nil, fmt.Errorf("no Engine named [%s] was found", name)
	}
}
