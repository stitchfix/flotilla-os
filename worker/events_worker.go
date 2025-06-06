package worker

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/utils"
	"gopkg.in/tomb.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	emrHistoryServer  string
	emrAppServer      map[string]string
	emrMetricsServer  string
	eksMetricsServer  string
	emrMaxPodEvents   int
	eksEngine         engine.Engine
	emrEngine         engine.Engine
	clusterManager    *engine.DynamicClusterManager
}

func (ew *eventsWorker) Initialize(conf config.Config, sm state.Manager, eksEngine engine.Engine, emrEngine engine.Engine, log flotillaLog.Logger, pollInterval time.Duration, qm queue.Manager, clusterManager *engine.DynamicClusterManager) error {
	ew.pollInterval = pollInterval
	ew.conf = conf
	ew.sm = sm
	ew.qm = qm
	ew.log = log
	ew.eksEngine = eksEngine
	ew.emrEngine = emrEngine
	eventsQueue, err := ew.qm.QurlFor(conf.GetString("eks_events_queue"), false)
	emrJobStatusQueue, err := ew.qm.QurlFor(conf.GetString("emr_job_status_queue"), false)
	ew.emrHistoryServer = conf.GetString("emr_history_server_uri")
	ew.emrAppServer = conf.GetStringMapString("emr_app_server_uri")
	ew.emrMetricsServer = conf.GetString("emr_metrics_server_uri")
	ew.eksMetricsServer = conf.GetString("eks_metrics_server_uri")
	ew.clusterManager = clusterManager
	if conf.IsSet("emr_max_attempt_count") {
		ew.emrMaxPodEvents = conf.GetInt("emr_max_pod_events")
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

	return nil
}

func (ew *eventsWorker) GetTomb() *tomb.Tomb {
	return &ew.t
}

func (ew *eventsWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-ew.t.Dying():
			_ = ew.log.Log("message", "An events worker was terminated")
			return nil
		default:
			loopCtx, span := utils.TraceJob(ctx, "events_worker.run_loop", "events_worker")
			ew.runOnce(loopCtx)
			ew.runOnceEMR(loopCtx)
			span.Finish()
			time.Sleep(ew.pollInterval)
		}
	}
}

func (ew *eventsWorker) runOnceEMR(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "events_worker.run_once_emr", "events_worker")
	defer span.Finish()
	emrEvent, err := ew.qm.ReceiveEMREvent(ew.emrJobStatusQueue)
	if err != nil {
		_ = ew.log.Log("message", "Error receiving EMR Events", "error", fmt.Sprintf("%+v", err))
		return
	}
	ew.processEventEMR(ctx, emrEvent)
}

func (ew *eventsWorker) processEventEMR(ctx context.Context, emrEvent state.EmrEvent) {
	if emrEvent.Detail == nil {
		return
	}

	emrJobId := emrEvent.Detail.ID
	run, err := ew.sm.GetRunByEMRJobId(ctx, *emrJobId)
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
			if run.StartedAt == nil || run.StartedAt.After(*run.FinishedAt) {
				run.StartedAt = run.QueuedAt
			}
			run.ExitReason = emrEvent.Detail.StateDetails
			// var events state.PodEvents
			// Pod Events are verbose and should be only stored for failed or running jobs.
			// run.PodEvents = &events
		case "RUNNING":
			run.Status = state.StatusRunning
			run.StartedAt = &timestamp
		case "FAILED":
			run.ExitCode = aws.Int64(-1)
			run.Status = state.StatusStopped
			run.FinishedAt = &timestamp
			if run.StartedAt == nil || run.StartedAt.After(*run.FinishedAt) {
				run.StartedAt = run.QueuedAt
			}

			run.ExitReason = aws.String("Job failed, please look at Driver Init and/or Driver Stdout logs.")

			if emrEvent.Detail != nil {
				if emrEvent.Detail.StateDetails != nil && !strings.Contains(*emrEvent.Detail.StateDetails, "JobRun failed. Please refer logs uploaded") {
					exitReason := strings.Replace(*emrEvent.Detail.StateDetails, "Please refer logs uploaded to S3/CloudWatch based on your monitoring configuration.", "", -1)
					run.ExitReason = aws.String(exitReason)
				} else {
					if emrEvent.Detail.FailureReason != nil && !strings.Contains(*emrEvent.Detail.FailureReason, "USER_ERROR") {
						exitReason := strings.Replace(*emrEvent.Detail.FailureReason, "Please refer logs uploaded to S3/CloudWatch based on your monitoring configuration.", "", -1)
						run.ExitReason = aws.String(exitReason)
					}
				}
			}

			if run.SparkExtension.DriverOOM != nil && *run.SparkExtension.DriverOOM == true {
				run.ExitReason = aws.String("Driver OOMKilled, retry with more driver memory.")
				run.ExitCode = aws.Int64(137)
			}

			if run.SparkExtension.ExecutorOOM != nil && *run.SparkExtension.ExecutorOOM == true {
				run.ExitReason = aws.String("Executor OOMKilled, retry with more executor memory.")
				run.ExitCode = aws.Int64(137)
			}

		case "SUBMITTED":
			run.Status = state.StatusPending
		}

		ew.setEMRMetricsUri(&run)
		_, err = ew.sm.UpdateRun(ctx, run.RunID, run)
		if err == nil {
			_ = emrEvent.Done()
		}
	}
}
func (ew *eventsWorker) runOnce(ctx context.Context) {
	ctx, span := utils.TraceJob(ctx, "events_worker.run_once_eks", "events_worker")
	defer span.Finish()
	kubernetesEvent, err := ew.qm.ReceiveKubernetesEvent(ew.queue)
	if err != nil {
		_ = ew.log.Log("message", "Error receiving Kubernetes Events", "error", fmt.Sprintf("%+v", err))
		return
	}
	ew.processEvent(ctx, kubernetesEvent)
}

func (ew *eventsWorker) processEMRPodEvents(ctx context.Context, kubernetesEvent state.KubernetesEvent) {
	if kubernetesEvent.InvolvedObject.Kind == "Pod" {
		// Skip events with empty cluster name
		if kubernetesEvent.InvolvedObject.Labels.ClusterName == "" {
			_ = kubernetesEvent.Done()
			return
		}
		var emrJobId *string = nil
		var sparkAppId *string = nil
		var driverServiceName *string = nil
		var executorOOM *bool = nil
		var driverOOM *bool = nil

		kClient, err := ew.clusterManager.GetKubernetesClient(kubernetesEvent.InvolvedObject.Labels.ClusterName)

		if err == nil {
			pod, err := kClient.CoreV1().Pods(kubernetesEvent.InvolvedObject.Namespace).Get(ctx, kubernetesEvent.InvolvedObject.Name, metav1.GetOptions{})
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

			if pod != nil {
				for _, container := range pod.Spec.Containers {
					for _, v := range container.Env {
						if v.Name == "SPARK_DRIVER_URL" {
							pat := regexp.MustCompile(`.*@(.*-svc).*`)
							matches := pat.FindAllStringSubmatch(v.Value, -1)
							for _, match := range matches {
								if len(match) == 2 {
									driverServiceName = &match[1]
								}
							}
						}
					}
				}

				if pod.Status.ContainerStatuses != nil && len(pod.Status.ContainerStatuses) > 0 {
					for _, containerStatus := range pod.Status.ContainerStatuses {
						if containerStatus.State.Terminated != nil {
							if containerStatus.State.Terminated.ExitCode == 137 {
								if strings.Contains(containerStatus.Name, "driver") {
									driverOOM = aws.Bool(true)
									_ = ew.log.Log("message", "Detected driver OOM",
										"container", containerStatus.Name)
								} else {
									executorOOM = aws.Bool(true)
									_ = ew.log.Log("message", "Detected executor OOM",
										"container", containerStatus.Name)
								}
							}
						}
					}
				}
			}
		}

		if emrJobId != nil {
			run, err := ew.sm.GetRunByEMRJobId(ctx, *emrJobId)
			if err == nil {
				_, span := utils.TraceJob(ctx, "flotilla.job.process_emr_pod_event", run.RunID)
				defer span.Finish()
				utils.TagJobRun(span, run)
				span.SetTag("emr.job_id", *emrJobId)
				layout := "2006-01-02T15:04:05Z"
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

				if executorOOM != nil && *executorOOM == true {
					run.SparkExtension.ExecutorOOM = executorOOM
				}
				if driverOOM != nil && *driverOOM == true {
					run.SparkExtension.DriverOOM = driverOOM
				}

				if sparkAppId != nil {
					sparkHistoryUri := fmt.Sprintf("%s/%s/jobs/", ew.emrHistoryServer, *sparkAppId)
					run.SparkExtension.SparkAppId = sparkAppId
					run.SparkExtension.HistoryUri = &sparkHistoryUri

					if driverServiceName != nil {
						appUri := ""
						if run.SparkExtension.SparkServerURI != nil {
							appUri = fmt.Sprintf("%s/job/%s", *run.SparkExtension.SparkServerURI, *driverServiceName)
						} else if serverURI, ok := ew.emrAppServer[run.ClusterName]; ok {
							appUri = fmt.Sprintf("%s/job/%s", serverURI, *driverServiceName)
						}

						if appUri != "" {
							run.SparkExtension.AppUri = &appUri
						}
					}
				}

				ew.setEMRMetricsUri(&run)

				run, err = ew.sm.UpdateRun(ctx, run.RunID, run)
				if err != nil {
					_ = ew.log.Log("message", "error saving kubernetes events", "emrJobId", emrJobId, "error", fmt.Sprintf("%+v", err))
					span.SetTag("error", true)
					span.SetTag("error.msg", err.Error())
				}

				if run.PodEvents != nil && len(*run.PodEvents) >= ew.emrMaxPodEvents {
					_ = ew.emrEngine.Terminate(ctx, run)
				}

			}
		}
		_ = kubernetesEvent.Done()
	}
}

func (ew *eventsWorker) setEMRMetricsUri(run *state.Run) {
	if run != nil && run.SparkExtension != nil && run.SparkExtension.SparkAppId != nil {
		// https://production-stitchfix.datadoghq.com/data-jobs?query=%40app_id%3Aspark-000000035ee16lm6uri
		metricsUri :=
			fmt.Sprintf("%s?query=%%40app_id%%3A%s",
				ew.emrMetricsServer,
				*run.SparkExtension.SparkAppId,
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
			fmt.Sprintf("%s&tpl_var_flotilla_run_id=%s&from_ts=%d&to_ts=%d&live=true",
				ew.eksMetricsServer,
				run.RunID,
				from,
				to,
			)
		run.MetricsUri = &metricsUri
	}
}

func (ew *eventsWorker) processEvent(ctx context.Context, kubernetesEvent state.KubernetesEvent) {
	runId := kubernetesEvent.InvolvedObject.Labels.JobName
	ctx, span := utils.TraceJob(ctx, "flotilla.job.process_event", runId)
	defer span.Finish()
	if strings.HasPrefix(runId, "eks-spark") || len(runId) == 0 {
		ew.processEMRPodEvents(ctx, kubernetesEvent)
		return
	}

	layout := "2020-08-31T17:27:50Z"
	timestamp, err := time.Parse(layout, kubernetesEvent.FirstTimestamp)

	if err != nil {
		timestamp = time.Now()
	}

	run, err := ew.sm.GetRun(ctx, runId)
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
		run, err = ew.sm.UpdateRun(ctx, runId, run)
		if err != nil {
			_ = ew.log.Log("message", "error saving kubernetes events", "run", runId, "error", fmt.Sprintf("%+v", err))
		} else {
			_ = kubernetesEvent.Done()
		}
	}
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
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
