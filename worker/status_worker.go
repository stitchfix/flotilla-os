package worker

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/clients/metrics"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type statusWorker struct {
	sm                       state.Manager
	ee                       engine.Engine
	conf                     config.Config
	log                      flotillaLog.Logger
	pollInterval             time.Duration
	t                        tomb.Tomb
	engine                   *string
	redisClient              *redis.Client
	workerId                 string
	exceptionExtractorClient *http.Client
	exceptionExtractorUrl    string
}

func (sw *statusWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string, qm queue.Manager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = ee
	sw.log = log
	sw.engine = engine
	sw.workerId = fmt.Sprintf("workerid:%d", rand.Int())

	if sw.conf.IsSet("eks.exception_extractor_url") {
		sw.exceptionExtractorClient = &http.Client{
			Timeout: time.Second * 5,
		}
		sw.exceptionExtractorUrl = sw.conf.GetString("eks.exception_extractor_url")
	}
	sw.setupRedisClient(conf)
	_ = sw.log.Log("message", "initialized a status worker", "engine", *engine)
	return nil
}

func (sw *statusWorker) setupRedisClient(conf config.Config) {
	if *sw.engine == state.EKSEngine {
		sw.redisClient = redis.NewClient(&redis.Options{Addr: conf.GetString("redis_address"), DB: conf.GetInt("redis_db")})
	}
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
				time.Sleep(sw.pollInterval)
			}
		}
	}
}

func (sw *statusWorker) runOnceEKS() {
	rl, err := sw.sm.ListRuns(1000, 0, "started_at", "asc", map[string][]string{
		"queued_at_since": {
			time.Now().AddDate(0, 0, -30).Format(time.RFC3339),
		},
		"status": {state.StatusNeedsRetry, state.StatusRunning, state.StatusQueued, state.StatusPending},
	}, nil, []string{state.EKSEngine})

	if err != nil {
		_ = sw.log.Log("message", "unable to receive runs", "error", fmt.Sprintf("%+v", err))
		return
	}
	runs := rl.Runs
	sw.processEKSRuns(runs)
}

func (sw *statusWorker) processEKSRuns(runs []state.Run) {
	var lockedRuns []state.Run
	for _, run := range runs {
		duration := time.Duration(30) * time.Second
		lock := sw.acquireLock(run, "status", duration)
		if lock {
			lockedRuns = append(lockedRuns, run)
		}
	}
	_ = metrics.Increment(metrics.StatusWorkerLockedRuns, []string{sw.workerId}, float64(len(lockedRuns)))
	for _, run := range lockedRuns {
		start := time.Now()
		_ = sw.log.Log("message", "launching go process eks run", "run", run.RunID)
		go sw.processEKSRun(run)
		_ = metrics.Timing(metrics.StatusWorkerProcessEKSRun, time.Since(start), []string{sw.workerId}, 1)
	}
}
func (sw *statusWorker) acquireLock(run state.Run, purpose string, expiration time.Duration) bool {
	start := time.Now()
	key := fmt.Sprintf("%s-%s", run.RunID, purpose)
	ttl, err := sw.redisClient.TTL(key).Result()
	if err == nil && ttl.Nanoseconds() < 0 {
		_, err = sw.redisClient.Del(key).Result()
	}
	set, err := sw.redisClient.SetNX(key, sw.workerId, expiration).Result()
	if err != nil {
		_ = sw.log.Log("message", "unable to set lock", "error", fmt.Sprintf("%+v", err))
		return true
	}
	_ = metrics.Timing(metrics.StatusWorkerAcquireLock, time.Since(start), []string{sw.workerId}, 1)
	return set
}

func (sw *statusWorker) processEKSRun(run state.Run) {
	_ = sw.log.Log("message", "process eks run", "run", run.RunID)
	reloadRun, err := sw.sm.GetRun(run.RunID)
	if err == nil && reloadRun.Status == state.StatusStopped {
		// Run was updated by another worker process.
		return
	}
	start := time.Now()
	updatedRunWithMetrics, _ := sw.ee.FetchPodMetrics(run)
	_ = metrics.Timing(metrics.StatusWorkerFetchPodMetrics, time.Since(start), []string{sw.workerId}, 1)

	start = time.Now()
	updatedRun, err := sw.ee.FetchUpdateStatus(updatedRunWithMetrics)
	if err != nil {
		_ = sw.log.Log("message", "fetch update status", "run", run.RunID, "error", fmt.Sprintf("%+v", err))
	}
	_ = metrics.Timing(metrics.StatusWorkerFetchUpdateStatus, time.Since(start), []string{sw.workerId}, 1)

	if err == nil {
		subRuns, err := sw.sm.ListRuns(1000, 0, "status", "desc", nil, map[string]string{"PARENT_FLOTILLA_RUN_ID": run.RunID}, state.Engines)
		if err == nil && subRuns.Total > 0 {
			var spawnedRuns state.SpawnedRuns
			for _, subRun := range subRuns.Runs {
				spawnedRuns = append(spawnedRuns, state.SpawnedRun{RunID: subRun.RunID})
			}
			updatedRun.SpawnedRuns = &spawnedRuns
		}
	}
	if err != nil {
		message := fmt.Sprintf("%+v", err)
		_ = sw.log.Log("message", "unable to receive eks runs", "error", message)

		minutesInQueue := time.Now().Sub(*run.QueuedAt).Minutes()
		if strings.Contains(message, "not found") && minutesInQueue > float64(30) {
			stoppedAt := time.Now()
			reason := "Job either timed out or not found on the EKS cluster."
			updatedRun.Status = state.StatusStopped
			updatedRun.FinishedAt = &stoppedAt
			updatedRun.ExitReason = &reason
			_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
		}

	} else {
		if run.Status != updatedRun.Status && (updatedRun.PodName == run.PodName) {
			_ = sw.log.Log("message", "updating eks run status", "pod", updatedRun.PodName, "status", updatedRun.Status, "exit_code", updatedRun.ExitCode)

			if updatedRun.ExitCode != nil {
				go sw.cleanupRun(run.RunID)
			}
			_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
			if err != nil {
				_ = sw.log.Log("message", "unable to save eks runs", "error", fmt.Sprintf("%+v", err))
			}

			if updatedRun.Status == state.StatusStopped {
				//TODO - move to a separate worker.
				//_ = sw.ee.Terminate(run)
			}
		} else {
			if updatedRun.MaxMemoryUsed != run.MaxMemoryUsed ||
				updatedRun.MaxCpuUsed != run.MaxCpuUsed ||
				updatedRun.Cpu != run.Cpu ||
				updatedRun.PodName != run.PodName ||
				updatedRun.Memory != run.Memory ||
				updatedRun.PodEvents != run.PodEvents ||
				updatedRun.SpawnedRuns != run.SpawnedRuns {
				_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
			}
		}
	}
}

func (sw *statusWorker) cleanupRun(runID string) {
	//Logs maybe delayed before being persisted to S3.
	time.Sleep(120 * time.Second)
	run, err := sw.sm.GetRun(runID)
	if err == nil {
		//Delete run from Kubernetes
		_ = sw.ee.Terminate(run)
	}
}

func (sw *statusWorker) extractExceptions(runID string) {
	//Logs maybe delayed before being persisted to S3.
	time.Sleep(60 * time.Second)
	run, err := sw.sm.GetRun(runID)
	if err == nil {
		jobUrl := fmt.Sprintf("%s/extract/%s", sw.exceptionExtractorUrl, run.RunID)
		res, err := sw.exceptionExtractorClient.Get(jobUrl)
		if err == nil && res != nil && res.Body != nil {
			body, err := ioutil.ReadAll(res.Body)
			if body != nil {
				defer res.Body.Close()
				runExceptions := state.RunExceptions{}
				err = json.Unmarshal(body, &runExceptions)
				if err == nil {
					run.RunExceptions = &runExceptions
				}
			}
			_, _ = sw.sm.UpdateRun(run.RunID, run)
		}
	}
}

func (sw *statusWorker) processEKSRunMetrics(run state.Run) {
	updatedRun, err := sw.ee.FetchPodMetrics(run)
	if err == nil {
		if updatedRun.MaxMemoryUsed != run.MaxMemoryUsed ||
			updatedRun.MaxCpuUsed != run.MaxCpuUsed {
			_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
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

		if sw.conf.GetString("flotilla_mode") != "dev" {
			_ = sw.log.Log("message", "Acking status update", "arn", update.TaskArn)
		}
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
			"env", env,
			"executable_id", update.ExecutableID,
			"executable_type", update.ExecutableType)
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
			"env", env,
			"executable_id", update.ExecutableID,
			"executable_type", update.ExecutableType)
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
