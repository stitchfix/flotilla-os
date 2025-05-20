package engine

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

// Engine defines the execution engine interface.
type Engine interface {
	Initialize(conf config.Config) error
	Execute(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) (state.Run, bool, error)
	Terminate(ctx context.Context, run state.Run) error
	Enqueue(ctx context.Context, run state.Run) error
	PollRuns(ctx context.Context) ([]RunReceipt, error)
	PollRunStatus(ctx context.Context) (state.Run, error)
	PollStatus(ctx context.Context) (RunReceipt, error)
	GetEvents(ctx context.Context, run state.Run) (state.PodEventList, error)
	FetchUpdateStatus(ctx context.Context, run state.Run) (state.Run, error)
	FetchPodMetrics(ctx context.Context, run state.Run) (state.Run, error)
	// Legacy methods from the ECS era. Here for backwards compatibility.
	Define(ctx context.Context, definition state.Definition) (state.Definition, error)
	Deregister(ctx context.Context, definition state.Definition) error
}

type RunReceipt struct {
	queue.RunReceipt
}

// NewExecutionEngine initializes and returns a new Engine
func NewExecutionEngine(conf config.Config, qm queue.Manager, name string, logger log.Logger, clusterManager *DynamicClusterManager, stateManager state.Manager) (Engine, error) {
	switch name {
	case state.EKSEngine:
		eksEng := &EKSExecutionEngine{qm: qm, log: logger, clusterManager: clusterManager, stateManager: stateManager}
		if err := eksEng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing EKSExecutionEngine")
		}
		return eksEng, nil
	case state.EKSSparkEngine:
		emrEng := &EMRExecutionEngine{sqsQueueManager: qm, log: logger, clusterManager: clusterManager, stateManager: stateManager}
		if err := emrEng.Initialize(conf); err != nil {
			return nil, errors.Wrap(err, "problem initializing EMRExecutionEngine")
		}
		return emrEng, nil
	default:
		return nil, fmt.Errorf("no Engine named [%s] was found", name)
	}
}
