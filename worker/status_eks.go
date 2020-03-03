package worker

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/stitchfix/flotilla-os/queue"
	"math/rand"
	"strings"
	"time"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
)

type statusEKSWorker struct {
	sm           state.Manager
	ee           engine.Engine
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	engine       *string
	redisClient  *redis.Client
	workerId     string
}

func (sw *statusEKSWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string, qm queue.Manager) error {
	sw.pollInterval = pollInterval
	sw.conf = conf
	sw.sm = sm
	sw.ee = ee
	sw.log = log
	sw.engine = engine
	sw.workerId = fmt.Sprintf("%d", rand.Int())
	sw.setupRedisClient(conf)
	_ = sw.log.Log("message", "initialized a status-eks worker", "engine", *engine)
	return nil
}

func (sw *statusEKSWorker) setupRedisClient(conf config.Config) {
	if *sw.engine == state.EKSEngine {
		sw.redisClient = redis.NewClient(&redis.Options{Addr: conf.GetString("redis_address"), DB: conf.GetInt("redis_db")})
	}
}

func (sw *statusEKSWorker) GetTomb() *tomb.Tomb {
	return &sw.t
}

//
// Run updates status of tasks
//
func (sw *statusEKSWorker) Run() error {
	for {
		select {
		case <-sw.t.Dying():
			_ = sw.log.Log("message", "A status worker was terminated")
			return nil
		default:
			if *sw.engine == state.EKSEngine {
				sw.runOnceEKS()
				time.Sleep(sw.pollInterval)
			}
		}
	}
}

func (sw *statusEKSWorker) runOnceEKS() {
	run, err := sw.ee.PollRunStatus()
	if err == nil {
		sw.processEKSRun(run)
	}
}

func (sw *statusEKSWorker) processEKSRun(run state.Run) {
	reloadRun, err := sw.sm.GetRun(run.RunID)
	if err != nil {
		return
	}

	if sw.acquireLock(run, "status", 10*time.Second) == false {
		return
	}
	_ = sw.log.Log("message", "valid run found", run.RunID)
	run = reloadRun
	updatedRun, err := sw.ee.FetchUpdateStatus(run)
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
		if run.Status != updatedRun.Status {
			_ = sw.log.Log("message", "updating eks run", "run", updatedRun.RunID, "status", updatedRun.Status)
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
				updatedRun.Memory != run.Memory ||
				updatedRun.PodEvents != run.PodEvents {
				_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
			}
		}
	}
}

func (sw *statusEKSWorker) processEKSRunMetrics(run state.Run) {
	updatedRun, err := sw.ee.FetchPodMetrics(run)
	if err == nil {
		if updatedRun.MaxMemoryUsed != run.MaxMemoryUsed ||
			updatedRun.MaxCpuUsed != run.MaxCpuUsed {
			_, err = sw.sm.UpdateRun(updatedRun.RunID, updatedRun)
		}
	}
}
func (sw *statusEKSWorker) acquireLock(run state.Run, purpose string, expiration time.Duration) bool {
	set, err := sw.redisClient.SetNX(fmt.Sprintf("%s-%s", run.RunID, purpose), sw.workerId, expiration).Result()
	if err != nil {
		// Turn off in dev mode; too noisy.
		if sw.conf.GetString("flotilla_mode") != "dev" {
			_ = sw.log.Log("message", "unable to set lock", "error", fmt.Sprintf("%+v", err))
		}
		return false
	}
	return set
}
