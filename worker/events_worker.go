package worker

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"gopkg.in/tomb.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"regexp"
	"strings"
	"time"
)

type eventsWorker struct {
	sm                state.Manager
	qm                queue.Manager
	conf              config.Config
	log               flotillaLog.Logger
	pollInterval      time.Duration
	t                 tomb.Tomb
	queue             string
	emrJobStatusQueue string
	s3Client          *s3.S3
	kClient           kubernetes.Clientset
	emrHistoryServer  string
	emrMetricsServer  string
	eksMetricsServer  string
	emrMaxPodEvents   int
	eksEngine         engine.Engine
	emrEngine         engine.Engine
}

func (ew *eventsWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager) error {
	ew.pollInterval = pollInterval
	ew.conf = conf
	ew.sm = sm
	ew.qm = qm
	ew.log = log
	ew.eksEngine = eksEngine
	ew.emrEngine = emrEngine
	eventsQueue, err := ew.qm.QurlFor(conf.GetString("eks.events_queue"), false)
	emrJobStatusQueue, err := ew.qm.QurlFor(conf.GetString("emr.job_status_queue"), false)
	ew.emrHistoryServer = conf.GetString("emr.history_server_uri")
	ew.emrMetricsServer = conf.GetString("emr.metrics_server_uri")
	ew.eksMetricsServer = conf.GetString("eks.metrics_server_uri")
	if conf.IsSet("emr.max_attempt_count") {
		ew.emrMaxPodEvents = conf.GetInt("emr.max_pod_events")
	} else {
		ew.emrMaxPodEvents = 20000
	}

	if err != nil {
		_ = ew.log.Log("message", "Error receiving Kubernetes Event queue", "error", fmt.Sprintf("%+v", err))
		return nil
	}
	ew.queue = eventsQueue
	ew.emrJobStatusQueue = emrJobStatusQueue
	_ = ew.qm.Initialize(ew.conf, "eks")

	clusterName := conf.GetStringSlice("eks.cluster_override")[0]

	filename := fmt.Sprintf("%s/%s", conf.GetString("eks.kubeconfig_basepath"), clusterName)
	clientConf, err := clientcmd.BuildConfigFromFlags("", filename)
	if err != nil {
		_ = ew.log.Log("message", "error initializing-eksEngine-clusters", "error", fmt.Sprintf("%+v", err))
		return err
	}
	kClient, err := kubernetes.NewForConfig(clientConf)
	if err != nil {
		_ = ew.log.Log("message", fmt.Sprintf("%+v", err))
		return err
	}
	ew.kClient = *kClient
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
			ew.runOnceEMR()
			time.Sleep(ew.pollInterval)
		}
	}
}

func (ew *eventsWorker) runOnceEMR() {
	emrEvent, err := ew.qm.ReceiveEMREvent(ew.emrJobStatusQueue)
	if err != nil {
		_ = ew.log.Log("message", "Error receiving EMR Events", "error", fmt.Sprintf("%+v", err))
		return
	}
	ew.processEventEMR(emrEvent)
}

func (ew *eventsWorker) processEventEMR(emrEvent state.EmrEvent) {
	if emrEvent.Detail == nil {
		return
	}

	emrJobId := emrEvent.Detail.ID
	run, err := ew.sm.GetRunByEMRJobId(*emrJobId)
	if err == nil {
		layout := "2020-08-31T17:27:50Z"
		timestamp, err := time.Parse(layout, *emrEvent.Time)
		if err != nil {
			timestamp = time.Now()
		}
		switch *emrEvent.Detail.State {
		case "COMPLETED":
			run.ExitCode = aws.Int64(0)
			run.Status = state.StatusStopped
			run.FinishedAt = &timestamp
			if run.StartedAt == nil {
				run.StartedAt = run.QueuedAt
			}
			run.ExitReason = emrEvent.Detail.StateDetails
			var events state.PodEvents
			// Pod Events are verbose and should be only stored for failed or running jobs.
			run.PodEvents = &events
		case "RUNNING":
			run.Status = state.StatusRunning
			run.StartedAt = &timestamp
		case "FAILED":
			run.ExitCode = aws.Int64(-1)
			run.Status = state.StatusStopped
			run.FinishedAt = &timestamp
			if run.StartedAt == nil {
				run.StartedAt = run.QueuedAt
			}
			run.ExitReason = emrEvent.Detail.FailureReason
		case "SUBMITTED":
			run.Status = state.StatusPending
		}

		ew.setEMRMetricsUri(&run)
		_, err = ew.sm.UpdateRun(run.RunID, run)
		if err == nil {
			_ = emrEvent.Done()
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
func (ew *eventsWorker) processEMRPodEvents(kubernetesEvent state.KubernetesEvent) {
	if kubernetesEvent.InvolvedObject.Kind == "Pod" {
		pod, err := ew.kClient.CoreV1().Pods(kubernetesEvent.InvolvedObject.Namespace).Get(kubernetesEvent.InvolvedObject.Name, metav1.GetOptions{})
		var emrJobId *string = nil
		var sparkAppId *string = nil
		if err == nil {
			for k, v := range pod.Labels {
				if emrJobId == nil && strings.Compare(k, "emr-containers.amazonaws.com/job.id") == 0 {
					emrJobId = aws.String(v)
				}
				if sparkAppId == nil && strings.Compare(k, "spark-app-selector") == 0 {
					sparkAppId = aws.String(v)
				}
				if sparkAppId != nil && emrJobId != nil {
					break
				}
			}
		}

		if emrJobId != nil {
			run, err := ew.sm.GetRunByEMRJobId(*emrJobId)
			if err == nil {
				layout := "2020-08-31T17:27:50Z"
				timestamp, err := time.Parse(layout, kubernetesEvent.FirstTimestamp)
				if err != nil {
					timestamp = time.Now()
				}

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

				if sparkAppId != nil {
					sparkHistoryUri := fmt.Sprintf("%s/%s/jobs", ew.emrHistoryServer, *sparkAppId)
					run.SparkExtension.SparkAppId = sparkAppId
					run.SparkExtension.HistoryUri = &sparkHistoryUri

				}
				ew.setEMRMetricsUri(&run)

				run, err = ew.sm.UpdateRun(run.RunID, run)
				if err != nil {
					_ = ew.log.Log("message", "error saving kubernetes events", "emrJobId", emrJobId, "error", fmt.Sprintf("%+v", err))
				}

				if run.PodEvents != nil && len(*run.PodEvents) >= ew.emrMaxPodEvents {
					_ = ew.emrEngine.Terminate(run)
				}

			}
		}
		_ = kubernetesEvent.Done()
	}
}

func (ew *eventsWorker) setEMRMetricsUri(run *state.Run) {
	if run != nil && run.SparkExtension != nil && run.SparkExtension.SparkAppId != nil {
		to := "now"

		if run.FinishedAt != nil {
			to = fmt.Sprintf("%d", run.FinishedAt.Add(time.Minute*1).UnixNano()/1000000)
		}

		from := time.Now().Add(-1*time.Minute*1).UnixNano() / 1000000
		if run.StartedAt != nil {
			from = run.StartedAt.Add(-1*time.Minute*1).UnixNano() / 1000000
		}

		metricsUri :=
			fmt.Sprintf("%svar-spark_app_selector=%s&from=%d&to=%s",
				ew.emrMetricsServer,
				*run.SparkExtension.SparkAppId,
				from,
				to,
			)

		run.MetricsUri = &metricsUri
	}
}

func (ew *eventsWorker) setEKSMetricsUri(run *state.Run) {
	if run != nil {
		to := time.Now().Add(1*time.Minute*1).UnixNano() / 1000000

		if run.FinishedAt != nil {
			to = run.FinishedAt.Add(time.Minute*1).UnixNano() / 1000000
		}

		from := time.Now().Add(-1*time.Minute*1).UnixNano() / 1000000
		if run.StartedAt != nil {
			from = run.StartedAt.Add(-1*time.Minute*1).UnixNano() / 1000000
		}

		metricsUri :=
			fmt.Sprintf("%svar-run_id=%s&from=%d&to=%d",
				ew.eksMetricsServer,
				run.RunID,
				from,
				to,
			)

		run.MetricsUri = &metricsUri
	}
}

func (ew *eventsWorker) processEvent(kubernetesEvent state.KubernetesEvent) {
	runId := kubernetesEvent.InvolvedObject.Labels.JobName
	if !strings.HasPrefix(runId, "eks") {
		ew.processEMRPodEvents(kubernetesEvent)
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
		ew.setEKSMetricsUri(&run)
		run, err = ew.sm.UpdateRun(runId, run)
		if err != nil {
			_ = ew.log.Log("message", "error saving kubernetes events", "run", runId, "error", fmt.Sprintf("%+v", err))
		} else {
			_ = kubernetesEvent.Done()
		}
	}
}

func (ew *eventsWorker) parsePodName(kubernetesEvent state.KubernetesEvent) (string, error) {
	expression := regexp.MustCompile(`(eks-\w+-\w+-\w+-\w+-\w+-\w+)`)
	matches := expression.FindStringSubmatch(kubernetesEvent.Message)
	if matches != nil && len(matches) >= 1 {
		return matches[0], nil
	}
	return "", errors.Errorf("no pod name found for [%s]", kubernetesEvent.Message)
}
