package worker

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"time"
)

type statusWorker struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
}

func (sw *statusWorker) Initialize(
	conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = ee
	sw.log = log
	return nil
}

//
// Run updates status of tasks
//
func (sw *statusWorker) Run() {
	for {
		sw.runOnce()
		time.Sleep(sw.pollInterval)
	}
}

func (sw *statusWorker) runOnce() {
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
	runs, err := sw.sm.ListRuns(1, 0, "started_at", "asc", map[string][]string{
		"task_arn": {taskArn},
	}, nil)
	if err != nil {
		return state.Run{}, errors.Wrapf(err, "problem finding run by task arn [%s]", taskArn)
	}
	if runs.Total > 0 && len(runs.Runs) > 0 {
		return runs.Runs[0], nil
	}
	return state.Run{}, errors.Errorf("no run found for [%s]", taskArn)
}
