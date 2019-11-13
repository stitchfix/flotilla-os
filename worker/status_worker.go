package worker

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
)

type statusWorker struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	engine       *string
}

func (sw *statusWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = ee
	sw.log = log
	sw.engine = engine
	sw.log.Log("message", "initialized a status worker", "engine", *engine)
	return nil
}

func (sw *statusWorker) GetTomb() *tomb.Tomb {
	return &sw.t
}

//
// Run updates status of tasks
//
func (sw *statusWorker) Run() error {
	for {
		select {
		case <-sw.t.Dying():
			sw.log.Log("message", "A status worker was terminated")
			return nil
		default:
			if *sw.engine == state.ECSEngine {
				sw.runOnceECS()
				time.Sleep(sw.pollInterval)
			}

			if *sw.engine == state.EKSEngine {
				sw.runOnceEKS()
				time.Sleep(time.Second * 15)
			}
		}
	}
}

func (sw *statusWorker) runOnceEKS() {
	rl, err := sw.sm.ListRuns(1000, 0, "status", "asc", map[string][]string{
		"queued_at_since": {
			time.Now().AddDate(0, 0, -1).Format(time.RFC3339),
		},
		"status": {state.StatusNeedsRetry, state.StatusRunning, state.StatusQueued, state.StatusPending},
	}, nil, []string{state.EKSEngine})

	if err != nil {
		sw.log.Log("message", "unable to receive runs", "error", fmt.Sprintf("%+v", err))
		return
	}

	for _, run := range rl.Runs {
		updatedRun, err := sw.ee.FetchUpdateStatus(run)
		if err != nil {
			sw.log.Log("message", "unable to receive runs", "error", fmt.Sprintf("%+v", err))
		} else {
			sw.log.Log("message", "updating run", "run", run.RunID)
			_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
			if err != nil {
				sw.log.Log("message", "unable to save runs", "error", fmt.Sprintf("%+v", err))
			}
		}
	}
}

func (sw *statusWorker) runOnceECS() {
	runReceipt, err := sw.ee.PollStatus()
	if err != nil {
		sw.log.Log("message", "unable to receive status message", "error", fmt.Sprintf("%+v", err))
		return
	}

	// Ensure update is in the env required, otherwise, ack without taking action
	update := runReceipt.Run
	if update != nil {
		//
		// Relies on the reserved env var, FLOTILLA_SERVER_MODE to ensure update
		// belongs to -this- mode of Flotilla
		//
		var serverMode string
		if update.Env != nil {
			for _, kv := range *update.Env {
				if kv.Name == "FLOTILLA_SERVER_MODE" {
					serverMode = kv.Value
				}
			}
		}

		shouldProcess := len(serverMode) > 0 && serverMode == sw.conf.GetString("flotilla_mode")
		if shouldProcess {
			run, err := sw.findRun(update.TaskArn)
			if err != nil {
				sw.log.Log("message", "unable to find run to apply update to", "error", fmt.Sprintf("%+v", err))
				return
			}

			_, err = sw.sm.UpdateRun(run.RunID, *update)
			if err != nil {
				sw.log.Log("message", "error applying status update", "run", run.RunID, "error", fmt.Sprintf("%+v", err))
				return
			}

			// emit status update event
			sw.logStatusUpdate(*update)
		}

		sw.log.Log("message", "Acking status update", "arn", update.TaskArn)
		if err = runReceipt.Done(); err != nil {
			sw.log.Log("message", "Acking status update failed", "arn", update.TaskArn, "error", fmt.Sprintf("%+v", err))
		}
	}
}

func (sw *statusWorker) logStatusUpdate(update state.Run) {
	var err error
	var startedAt, finishedAt time.Time
	var env state.EnvList

	if update.StartedAt != nil {
		startedAt = *update.StartedAt
	}

	if update.FinishedAt != nil {
		finishedAt = *update.FinishedAt
	}

	if update.Env != nil {
		env = *update.Env
	}

	if update.ExitCode != nil {
		err = sw.log.Event("eventClassName", "FlotillaTaskStatus",
			"run_id", update.RunID,
			"task_arn", update.TaskArn,
			"definition_id", update.DefinitionID,
			"alias", update.Alias,
			"image", update.Image,
			"cluster_name", update.ClusterName,
			"exit_code", *update.ExitCode,
			"status", update.Status,
			"started_at", startedAt,
			"finished_at", finishedAt,
			"instance_id", update.InstanceID,
			"instance_dns_name", update.InstanceDNSName,
			"group_name", update.GroupName,
			"user", update.User,
			"task_type", update.TaskType,
			"env", env)
	} else {
		err = sw.log.Event("eventClassName", "FlotillaTaskStatus",
			"run_id", update.RunID,
			"task_arn", update.TaskArn,
			"definition_id", update.DefinitionID,
			"alias", update.Alias,
			"image", update.Image,
			"cluster_name", update.ClusterName,
			"status", update.Status,
			"started_at", startedAt,
			"finished_at", finishedAt,
			"instance_id", update.InstanceID,
			"instance_dns_name", update.InstanceDNSName,
			"group_name", update.GroupName,
			"user", update.User,
			"task_type", update.TaskType,
			"env", env)
	}

	if err != nil {
		sw.log.Log("message", "Failed to emit status event", "run_id", update.RunID, "error", err.Error())
	}
}

func (sw *statusWorker) findRun(taskArn string) (state.Run, error) {
	var engines []string
	if sw.engine != nil {
		engines = []string{*sw.engine}
	} else {
		engines = nil
	}

	runs, err := sw.sm.ListRuns(1, 0, "started_at", "asc", map[string][]string{
		"task_arn": {taskArn},
	}, nil, engines)
	if err != nil {
		return state.Run{}, errors.Wrapf(err, "problem finding run by task arn [%s]", taskArn)
	}
	if runs.Total > 0 && len(runs.Runs) > 0 {
		return runs.Runs[0], nil
	}
	return state.Run{}, errors.Errorf("no run found for [%s]", taskArn)
}
