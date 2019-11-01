package engine

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

//
// Engine defines the execution engine interface.
//
type Engine interface {
	Initialize(conf config.Config) error
	Execute(definition state.Definition, run state.Run) (state.Run, bool, error)
	Define(definition state.Definition) (state.Definition, error)
	Deregister(definition state.Definition) error
	Terminate(run state.Run) error
	Enqueue(run state.Run) error
	PollRuns() ([]RunReceipt, error)
	PollStatus() (RunReceipt, error)
	Get(run state.Run) (state.Run, error)
}

type RunReceipt struct {
	queue.RunReceipt
}

//
// NewExecutionEngine initializes and returns a new Engine
//
func NewExecutionEngine(conf config.Config, qm queue.Manager, log flotillaLog.Logger) (Engine, error) {
	name := "ecs"
	if conf.IsSet("execution_engine") {
		name = conf.GetString("execution_engine")
	}

	switch name {
	case "ecs":
		eng := &ECSExecutionEngine{qm: qm}
		if err := eng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing ECSExecutionEngine")
		}
		return eng, nil
	case "eks":
		eng := &EKSExecutionEngine{qm: qm, log: log}
		if err := eng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing EKSExecutionEngine")
		}
		return eng, nil
	default:
		return nil, fmt.Errorf("no Engine named [%s] was found", name)
	}
}
