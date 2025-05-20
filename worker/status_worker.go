package worker

import (
	"context"
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
	"github.com/stitchfix/flotilla-os/utils"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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
	clusterManager           *engine.DynamicClusterManager
}

func (sw *statusWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager, clusterManager *engine.DynamicClusterManager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = eksEngine
	sw.log = log
	sw.workerId = fmt.Sprintf("workerid:%d", rand.Int())
	sw.engine = &state.EKSEngine
	sw.emrEngine = emrEngine
	sw.clusterManager = clusterManager
	if sw.conf.IsSet("eks_exception_extractor_url") {
		sw.exceptionExtractorClient = &http.Client{
			Timeout: time.Second * 5,
		}
		sw.exceptionExtractorUrl = sw.conf.GetString("eks_exception_extractor_url")
	}
	sw.redisClient, _ = utils.SetupRedisClient(conf)
	_ = sw.log.Log("message", "initialized a status worker")
	return nil
}

func (sw *statusWorker) GetTomb() *tomb.Tomb {
	return &sw.t
}

// Run updates status of tasks
func (sw *statusWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-sw.t.Dying():
			sw.log.Log("message", "A status worker was terminated")
			return nil
		default:
			if *sw.engine == state.EKSEngine {
				sw.runOnceEKS(ctx)
				sw.runTimeouts(ctx)
			}
			time.Sleep(sw.pollInterval)
		}
	}
}

func (sw *statusWorker) runTimeouts(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "status_worker.run_timeouts", sw.workerId)
	defer span.Finish()
	rl, err := sw.sm.ListRuns(ctx, 1000, 0, "started_at", "asc", map[string][]string{
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
	ctx := context.Background()
	span, ctx := tracer.StartSpanFromContext(ctx, "flotilla.job.timeout_check")
	defer span.Finish()
	span.SetTag("timeout_check.run_count", len(runs))
	timeoutCount := 0
	for _, run := range runs {
		if run.StartedAt != nil && run.ActiveDeadlineSeconds != nil {
			runningDuration := time.Now().Sub(*run.StartedAt)
			if int64(runningDuration.Seconds()) > *run.ActiveDeadlineSeconds {
				timeoutCount++
				timeoutCtx, childSpan := utils.TraceJob(ctx, "flotilla.job.timeout", run.RunID)
				utils.TagJobRun(childSpan, run)
				if run.Engine != nil && *run.Engine == state.EKSSparkEngine {
					_ = sw.emrEngine.Terminate(timeoutCtx, run)
				} else {
					_ = sw.ee.Terminate(timeoutCtx, run)
				}

				exitCode := int64(1)
				finishedAt := time.Now()
				_, _ = sw.sm.UpdateRun(ctx, run.RunID, state.Run{
					Status:     state.StatusStopped,
					ExitReason: aws.String(fmt.Sprintf("JobRun exceeded specified timeout of %v seconds", *run.ActiveDeadlineSeconds)),
					ExitCode:   &exitCode,
					FinishedAt: &finishedAt,
				})
				childSpan.Finish()
			}
		}
	}
	span.SetTag("timeout_check.timeout_count", timeoutCount)
}

func (sw *statusWorker) runOnceEKS(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "status_worker.run_once_eks", sw.workerId)
	defer span.Finish()
	rl, err := sw.sm.ListRuns(ctx, 1000, 0, "started_at", "asc", map[string][]string{
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
	sw.processEKSRuns(ctx, runs)
}

func (sw *statusWorker) processEKSRuns(ctx context.Context, runs []state.Run) {
	ctx, span := utils.TraceJob(ctx, "status_worker.process_eks_runs", sw.workerId)
	defer span.Finish()

	var lockedRuns []state.Run

	for _, run := range runs {
		_, lockSpan := utils.TraceJob(ctx, "status_worker.acquire_lock", run.RunID)

		duration := 45 * time.Second
		locked := sw.acquireLock(run, "status", duration)
		if locked {
			lockedRuns = append(lockedRuns, run)
			lockSpan.SetTag("lock.status", "acquired")
		} else {
			lockSpan.SetTag("lock.status", "not_acquired")
		}

		lockSpan.Finish()
	}

	_ = metrics.Increment(metrics.StatusWorkerLockedRuns, []string{sw.workerId}, float64(len(lockedRuns)))

	for _, run := range lockedRuns {
		runCopy := run
		go func() {
			runCtx, runSpan := utils.TraceJob(ctx, "flotilla.job.status_check", runCopy.RunID)
			defer runSpan.Finish()

			utils.TagJobRun(runSpan, runCopy)

			start := time.Now()
			sw.processEKSRun(runCtx, runCopy)
			_ = metrics.Timing(metrics.StatusWorkerProcessEKSRun, time.Since(start), []string{sw.workerId}, 1)
		}()
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

func (sw *statusWorker) processEKSRun(ctx context.Context, run state.Run) {
	ctx, span := utils.TraceJob(ctx, "flotilla.job.status_check", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	reloadRun, err := sw.sm.GetRun(ctx, run.RunID)
	if err == nil && reloadRun.Status == state.StatusStopped {
		// Run was updated by another worker process.
		return
	}
	start := time.Now()
	if reloadRun.Status == state.StatusQueued {
		queuedDuration := time.Since(*reloadRun.QueuedAt)
		if queuedDuration < 10*time.Second {
			return
		}
	}

	start = time.Now()
	statusCtx, statusSpan := utils.TraceJob(ctx, "flotilla.job.fetch_update_status", reloadRun.RunID)
	defer statusSpan.Finish()
	utils.TagJobRun(statusSpan, reloadRun)
	statusSpan.SetTag("cluster_name", reloadRun.ClusterName)

	updatedRun, err := sw.ee.FetchUpdateStatus(statusCtx, reloadRun)
	if err != nil {
		_ = sw.log.Log("message", "fetch update status", "run", run.RunID, "error", fmt.Sprintf("%+v", err))

		if strings.Contains(err.Error(), "not found") {
			if run.Status == state.StatusPending || run.Status == state.StatusQueued {
				statusSpan.SetTag("error.expected", true)
				statusSpan.SetTag("error", false)
			}
		}
	}
	_ = metrics.Timing(metrics.StatusWorkerFetchUpdateStatus, time.Since(start), []string{sw.workerId}, 1)

	if err == nil {
		subRuns, err := sw.sm.ListRuns(ctx, 1000, 0, "status", "desc", nil, map[string]string{"PARENT_FLOTILLA_RUN_ID": run.RunID}, state.Engines)
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
			_, err = sw.sm.UpdateRun(ctx, updatedRun.RunID, updatedRun)
		}

	} else {
		fullUpdate := false

		if run.PodName != nil {
			if *run.PodName == *updatedRun.PodName && run.Status != updatedRun.Status {
				fullUpdate = true
			}
		}

		if fullUpdate {
			sw.logStatusUpdate(updatedRun)
			if updatedRun.ExitCode != nil {
				go sw.cleanupRun(ctx, run.RunID)
			}
			_, err = sw.sm.UpdateRun(ctx, updatedRun.RunID, updatedRun)
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
				_, err = sw.sm.UpdateRun(ctx, updatedRun.RunID, updatedRun)
			}
		}
	}
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	} else if updatedRun.Status != run.Status {
		span.SetTag("job.status_change", fmt.Sprintf("%s->%s", run.Status, updatedRun.Status))
		utils.TagJobRun(span, updatedRun)
	}
}

func (sw *statusWorker) cleanupRun(ctx context.Context, runID string) {
	ctx, span := utils.TraceJob(ctx, "flotilla.job.cleanup", runID)
	defer span.Finish()

	defer span.Finish()
	//Logs maybe delayed before being persisted to S3.
	span.SetTag("cleanup.delay_start", time.Now().Unix())
	time.Sleep(120 * time.Second)
	span.SetTag("cleanup.delay_end", time.Now().Unix())
	run, err := sw.sm.GetRun(ctx, runID)
	if err == nil {
		//Delete run from Kubernetes
		_ = sw.ee.Terminate(ctx, run)
	}
}

func (sw *statusWorker) extractExceptions(ctx context.Context, runID string) {
	ctx, span := utils.TraceJob(ctx, "flotilla.job.extract_exceptions", runID)
	defer span.Finish()

	time.Sleep(60 * time.Second)
	run, err := sw.sm.GetRun(ctx, runID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return
	}
	jobUrl := fmt.Sprintf("%s/extract/%s", sw.exceptionExtractorUrl, run.RunID)
	res, err := sw.exceptionExtractorClient.Get(jobUrl)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		span.SetTag("error", true)
		return
	}
	var runExceptions state.RunExceptions
	if err := json.Unmarshal(body, &runExceptions); err == nil {
		run.RunExceptions = &runExceptions
		_, _ = sw.sm.UpdateRun(ctx, run.RunID, run)
	}
}

func (sw *statusWorker) processEKSRunMetrics(ctx context.Context, run state.Run) {
	ctx, span := utils.TraceJob(ctx, "flotilla.job.metrics_check", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	updatedRun, err := sw.ee.FetchPodMetrics(ctx, run)
	if err == nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		if updatedRun.MaxMemoryUsed != run.MaxMemoryUsed ||
			updatedRun.MaxCpuUsed != run.MaxCpuUsed {
			_, err = sw.sm.UpdateRun(ctx, updatedRun.RunID, updatedRun)
		}
	}
}

func (sw *statusWorker) logStatusUpdate(update state.Run) {
	var err error
	var startedAt, finishedAt time.Time
	var duration float64
	var env state.EnvList
	var command string

	if update.StartedAt != nil {
		startedAt = *update.StartedAt
		duration = time.Now().Sub(startedAt).Seconds()
	}

	if update.FinishedAt != nil {
		finishedAt = *update.FinishedAt
		duration = finishedAt.Sub(startedAt).Seconds()
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
			"duration", duration,
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
			"duration", duration,
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

func (sw *statusWorker) findRun(ctx context.Context, taskArn string) (state.Run, error) {
	ctx, span := utils.TraceJob(ctx, "status_worker.find_run", taskArn)
	defer span.Finish()

	var engines []string
	if sw.engine != nil {
		engines = []string{*sw.engine}
	}

	runs, err := sw.sm.ListRuns(ctx, 1, 0, "started_at", "asc", map[string][]string{
		"task_arn": {taskArn},
	}, nil, engines)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return state.Run{}, errors.Wrapf(err, "problem finding run by task arn [%s]", taskArn)
	}
	if runs.Total > 0 && len(runs.Runs) > 0 {
		return runs.Runs[0], nil
	}
	span.SetTag("not_found", true)
	return state.Run{}, errors.Errorf("no run found for [%s]", taskArn)
}
