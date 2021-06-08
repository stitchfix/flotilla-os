package worker

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
	"regexp"
	"strings"
	"time"
)

type eventsWorker struct {
	sm           state.Manager
	qm           queue.Manager
	conf         config.Config
	log          flotillaLog.Logger
	pollInterval time.Duration
	t            tomb.Tomb
	queue        string
	engine       *string
	s3Client     *s3.S3
}

func (ew *eventsWorker) Initialize(conf config.Config, sm state.Manager, ee engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, engine *string, qm queue.Manager) error {
	ew.pollInterval = pollInterval
	ew.conf = conf
	ew.sm = sm
	ew.qm = qm
	ew.log = log
	ew.engine = engine
	eventsQueue, err := ew.qm.QurlFor(conf.GetString("k8s.events_queue"), false)

	if err != nil {
		_ = ew.log.Log("message", "Error receiving Kubernetes Event queue", "error", fmt.Sprintf("%+v", err))
		return nil
	}

	ew.queue = eventsQueue
	_ = ew.qm.Initialize(ew.conf, "k8s")

	return nil
}

func (ew *eventsWorker) GetTomb() *tomb.Tomb {
	return &ew.t
}

func (ew *eventsWorker) Run() error {
	for {
		select {
		case <-ew.t.Dying():
			_ = ew.log.Log("message", "A CloudTrail worker was terminated")
			return nil
		default:
			ew.runOnce()
			time.Sleep(ew.pollInterval)
		}
	}
}

func (ew *eventsWorker) runOnce() {
	kubernetesEvent, err := ew.qm.ReceiveKubernetesEvent(ew.queue)
	if err != nil {
		_ = ew.log.Log("message", "Error receiving Kubernetes Events", "error", fmt.Sprintf("%+v", err))
		return
	}
	ew.processEvent(kubernetesEvent)
}

func (ew *eventsWorker) processEvent(kubernetesEvent state.KubernetesEvent) {
	runId := kubernetesEvent.InvolvedObject.Labels.JobName
	if !strings.HasPrefix(runId, "k8s") {
		return
	}

	layout := "2020-08-31T17:27:50Z"
	timestamp, err := time.Parse(layout, kubernetesEvent.FirstTimestamp)

	if err != nil {
		timestamp = time.Now()
	}

	run, err := ew.sm.GetRun(runId)
	if err == nil {
		event := state.PodEvent{
			Timestamp:    &timestamp,
			EventType:    kubernetesEvent.Type,
			Reason:       kubernetesEvent.Reason,
			SourceObject: kubernetesEvent.InvolvedObject.Name,
			Message:      kubernetesEvent.Message,
		}

		var events state.PodEvents
		if run.PodEvents != nil {
			events = append(*run.PodEvents, event)
		} else {
			events = state.PodEvents{event}
		}
		run.PodEvents = &events
		if kubernetesEvent.Reason == "Scheduled" {
			podName, err := ew.parsePodName(kubernetesEvent)
			if err == nil {
				run.PodName = &podName
			}
		}

		if kubernetesEvent.Reason == "DeadlineExceeded" {
			run.ExitReason = &kubernetesEvent.Message
			exitCode := int64(124)
			run.ExitCode = &exitCode
			run.Status = state.StatusStopped
			run.StartedAt = run.QueuedAt
			run.FinishedAt = &timestamp
		}

		if kubernetesEvent.Reason == "Completed" {
			run.ExitReason = &kubernetesEvent.Message
			exitCode := int64(0)
			run.ExitCode = &exitCode
			run.Status = state.StatusStopped
			run.StartedAt = run.QueuedAt
			run.FinishedAt = &timestamp
		}
		run, err = ew.sm.UpdateRun(runId, run)
		if err != nil {
			_ = ew.log.Log("message", "error saving kubernetes events", "run", runId, "error", fmt.Sprintf("%+v", err))
		} else {
			_ = kubernetesEvent.Done()
		}
	}
}

func (ew *eventsWorker) parsePodName(kubernetesEvent state.KubernetesEvent) (string, error) {
	expression := regexp.MustCompile(`(k8s-\w+-\w+-\w+-\w+-\w+-\w+)`)
	matches := expression.FindStringSubmatch(kubernetesEvent.Message)
	if matches != nil && len(matches) >= 1 {
		return matches[0], nil
	}
	return "", errors.Errorf("no pod name found for [%s]", kubernetesEvent.Message)
}
