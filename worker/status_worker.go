package worker

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
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
	emrEngine                engine.Engine
}

func (sw *statusWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = eksEngine
	sw.log = log
	sw.workerId = fmt.Sprintf("workerid:%d", rand.Int())
	sw.engine = &state.EKSEngine
	sw.emrEngine = emrEngine
	if sw.conf.IsSet("eks_exception_extractor_url") {
		sw.exceptionExtractorClient = &http.Client{
			Timeout: time.Second * 5,
		}
		sw.exceptionExtractorUrl = sw.conf.GetString("eks_exception_extractor_url")
	}
	sw.setupRedisClient(conf)
	_ = sw.log.Log("message", "initialized a status worker")
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
			if *sw.engine == state.EKSEngine {
				sw.runOnceEKS()
				sw.runTimeouts()
				time.Sleep(sw.pollInterval)
			}
		}
	}
}

func (sw *statusWorker) runTimeouts() {
	rl, err := sw.sm.ListRuns(1000, 0, "started_at", "asc", map[string][]string{
		"queued_at_since": {
			time.Now().AddDate(0, 0, -300).Format(time.RFC3339),
		},
		"task_type": {state.DefaultTaskType},
		"status":    {state.StatusNeedsRetry, state.StatusRunning, state.StatusQueued, state.StatusPending},
	}, nil, state.Engines)

	if err != nil {
		_ = sw.log.Log("message", "unable to receive runs", "error", fmt.Sprintf("%+v", err))
		return
	}
	runs := rl.Runs
	sw.processTimeouts(runs)
}

func (sw *statusWorker) processTimeouts(runs []state.Run) {
	for _, run := range runs {
		if run.StartedAt != nil && run.ActiveDeadlineSeconds != nil {
			runningDuration := time.Now().Sub(*run.StartedAt)
			if int64(runningDuration.Seconds()) > *run.ActiveDeadlineSeconds {
				if run.Engine != nil && *run.Engine == state.EKSSparkEngine {
					_ = sw.emrEngine.Terminate(run)
				} else {
					_ = sw.ee.Terminate(run)
				}

				exitCode := int64(1)
				finishedAt := time.Now()
				_, _ = sw.sm.UpdateRun(run.RunID, state.Run{
					Status:     state.StatusStopped,
					ExitReason: aws.String(fmt.Sprintf("JobRun exceeded specified timeout of %v seconds", *run.ActiveDeadlineSeconds)),
					ExitCode:   &exitCode,
					FinishedAt: &finishedAt,
				})
			}
		}
	}
}

func (sw *statusWorker) runOnceEKS() {
	rl, err := sw.sm.ListRuns(1000, 0, "started_at", "asc", map[string][]string{
		"queued_at_since": {
			time.Now().AddDate(0, 0, -300).Format(time.RFC3339),
		},
		"task_type": {state.DefaultTaskType},
		"status":    {state.StatusNeedsRetry, state.StatusRunning, state.StatusQueued, state.StatusPending},
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
		duration := time.Duration(45) * time.Second
		lock := sw.acquireLock(run, "status", duration)
		if lock {
			lockedRuns = append(lockedRuns, run)
		}
	}
	_ = metrics.Increment(metrics.StatusWorkerLockedRuns, []string{sw.workerId}, float64(len(lockedRuns)))
	for _, run := range lockedRuns {
		start := time.Now()
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
	reloadRun, err := sw.sm.GetRun(run.RunID)
	if err == nil && reloadRun.Status == state.StatusStopped {
		// Run was updated by another worker process.
		return
	}
	start := time.Now()

	start = time.Now()
	updatedRun, err := sw.ee.FetchUpdateStatus(reloadRun)
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
			sw.logStatusUpdate(updatedRun)
			if updatedRun.ExitCode != nil {
				go sw.cleanupRun(run.RunID)
			}
			_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
			if err != nil {
				_ = sw.log.Log("message", "unable to save eks runs", "error", fmt.Sprintf("%+v", err))
			}

			if updatedRun.Status == state.StatusStopped {
				//TODO - move to a separate worker.
				//_ = sw.eksEngine.Terminate(run)
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

func (sw *statusWorker) logStatusUpdate(update state.Run) {
	var err error
	var startedAt, finishedAt time.Time
	var env state.EnvList
	var command string

	if update.StartedAt != nil {
		startedAt = *update.StartedAt
	}

	if update.FinishedAt != nil {
		finishedAt = *update.FinishedAt
	}

	if update.Env != nil {
		env = *update.Env
	}

	if update.Command != nil {
		command = *update.Command
	}

	if update.ExitCode != nil {
		err = sw.log.Event("eventClassName", "FlotillaTaskStatus",
			"run_id", update.RunID,
			"definition_id", update.DefinitionID,
			"alias", update.Alias,
			"image", update.Image,
			"cluster_name", update.ClusterName,
			"command", command,
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
			"definition_id", update.DefinitionID,
			"alias", update.Alias,
			"image", update.Image,
			"cluster_name", update.ClusterName,
			"command", command,
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
