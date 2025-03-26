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
	ctx := context.Background()
	if kubernetesEvent.InvolvedObject.Kind == "Pod" {
		fmt.Printf("Processing pod event: %s, reason: %s, pod name: %s\n",
			kubernetesEvent.Type,
			kubernetesEvent.Reason,
			kubernetesEvent.InvolvedObject.Name)

		var emrJobId *string = nil
		var sparkAppId *string = nil
		var driverServiceName *string = nil
		var executorOOM *bool = nil
		var driverOOM *bool = nil

		kClient, err := ew.clusterManager.GetKubernetesClient(kubernetesEvent.InvolvedObject.Labels.ClusterName)
		if err != nil {
			fmt.Printf("Error getting Kubernetes client for cluster %s: %v\n",
				kubernetesEvent.InvolvedObject.Labels.ClusterName, err)
		} else {
			pod, err := kClient.CoreV1().Pods(kubernetesEvent.InvolvedObject.Namespace).Get(ctx, kubernetesEvent.InvolvedObject.Name, metav1.GetOptions{})
			if err != nil {
				fmt.Printf("Error getting pod %s: %v\n", kubernetesEvent.InvolvedObject.Name, err)
			} else {
				fmt.Printf("Pod phase: %s, pod name: %s\n", pod.Status.Phase, pod.Name)

				// Log labels
				fmt.Println("Pod labels:")
				for k, v := range pod.Labels {
					fmt.Printf("  %s = %s\n", k, v)
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

				// Check if containers are initialized
				fmt.Printf("Pod has %d containers\n", len(pod.Spec.Containers))
				containersReady := true
				if pod.Status.ContainerStatuses != nil {
					fmt.Printf("Container status count: %d\n", len(pod.Status.ContainerStatuses))
					for _, status := range pod.Status.ContainerStatuses {
						ready := status.Ready
						fmt.Printf("Container %s ready: %v\n", status.Name, ready)
						containersReady = containersReady && ready
					}
				} else {
					fmt.Println("No container statuses available yet")
					containersReady = false
				}

				fmt.Printf("All containers ready: %v\n", containersReady)

				// Continue with looking for driver URL
				if pod.Spec.Containers != nil && len(pod.Spec.Containers) > 0 {
					for i, container := range pod.Spec.Containers {
						fmt.Printf("Checking container %d: %s\n", i, container.Name)
						fmt.Printf("Container has %d env vars\n", len(container.Env))

						foundDriverURL := false
						for _, v := range container.Env {
							if v.Name == "SPARK_DRIVER_URL" {
								foundDriverURL = true
								fmt.Printf("Found SPARK_DRIVER_URL: %s\n", v.Value)

								pat := regexp.MustCompile(`.*@(.*-svc).*`)
								matches := pat.FindAllStringSubmatch(v.Value, -1)

								fmt.Printf("Regex found %d matches\n", len(matches))

								for _, match := range matches {
									fmt.Printf("Match has %d groups\n", len(match))
									for j, group := range match {
										fmt.Printf("Group %d: %s\n", j, group)
									}

									if len(match) == 2 {
										driverServiceName = &match[1]
										fmt.Printf("Set driver service name to: %s\n", *driverServiceName)
									}
								}
							}
						}

						if !foundDriverURL {
							fmt.Println("SPARK_DRIVER_URL environment variable not found")
						}
					}
				}

				if pod.Status.ContainerStatuses != nil && len(pod.Status.ContainerStatuses) > 0 {
					fmt.Println("Checking container termination status")
					for _, containerStatus := range pod.Status.ContainerStatuses {
						fmt.Printf("Container %s state: %+v\n", containerStatus.Name, containerStatus.State)
						if containerStatus.State.Terminated != nil {
							fmt.Printf("Container %s terminated with exit code: %d\n",
								containerStatus.Name, containerStatus.State.Terminated.ExitCode)
							if containerStatus.State.Terminated.ExitCode == 137 {
								if strings.Contains(containerStatus.Name, "driver") {
									driverOOM = aws.Bool(true)
									fmt.Println("Detected driver OOM")
								} else {
									executorOOM = aws.Bool(true)
									fmt.Println("Detected executor OOM")
								}
							}
						}
					}
				}
			}
		}

		fmt.Printf("Found EMR job ID: %v\n", emrJobId != nil)
		if emrJobId != nil {
			fmt.Printf("EMR job ID value: %s\n", *emrJobId)
		}

		fmt.Printf("Found Spark App ID: %v\n", sparkAppId != nil)
		fmt.Printf("Found driver service name: %v\n", driverServiceName != nil)

		if emrJobId != nil {
			fmt.Printf("Looking up run for EMR job ID: %s\n", *emrJobId)
			run, err := ew.sm.GetRunByEMRJobId(*emrJobId)
			if err != nil {
				fmt.Printf("Error getting run for EMR job ID %s: %v\n", *emrJobId, err)
			} else {
				fmt.Printf("Found run %s for EMR job ID %s\n", run.RunID, *emrJobId)

				layout := "2020-08-31T17:27:50Z"
				timestamp, err := time.Parse(layout, kubernetesEvent.FirstTimestamp)
				if err != nil {
					fmt.Printf("Error parsing timestamp: %v, using current time\n", err)
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
					fmt.Printf("Appending to existing %d pod events\n", len(*run.PodEvents))
				} else {
					events = state.PodEvents{event}
					fmt.Println("Creating new pod events array")
				}
				run.PodEvents = &events

				if executorOOM != nil && *executorOOM == true {
					fmt.Println("Setting executor OOM flag")
					run.SparkExtension.ExecutorOOM = executorOOM
				}
				if driverOOM != nil && *driverOOM == true {
					fmt.Println("Setting driver OOM flag")
					run.SparkExtension.DriverOOM = driverOOM
				}

				if sparkAppId != nil {
					fmt.Printf("Setting spark app ID: %s\n", *sparkAppId)
					sparkHistoryUri := fmt.Sprintf("%s/%s/jobs/", ew.emrHistoryServer, *sparkAppId)
					run.SparkExtension.SparkAppId = sparkAppId
					run.SparkExtension.HistoryUri = &sparkHistoryUri
					fmt.Printf("Set history URI: %s\n", sparkHistoryUri)

					if driverServiceName != nil {
						fmt.Printf("Setting app URI for driver service: %s\n", *driverServiceName)
						appUri := ""
						if run.SparkExtension.SparkServerURI != nil {
							fmt.Printf("Using SparkServerURI: %s\n", *run.SparkExtension.SparkServerURI)
							appUri = fmt.Sprintf("%s/job/%s", *run.SparkExtension.SparkServerURI, *driverServiceName)
						} else if _, ok := ew.emrAppServer[run.ClusterName]; ok {
							fmt.Printf("Using emrAppServer for cluster %s: %s\n",
								run.ClusterName, ew.emrAppServer[run.ClusterName])
							appUri = fmt.Sprintf("%s/job/%s", ew.emrAppServer[run.ClusterName], *driverServiceName)
						} else {
							fmt.Printf("No app server URL found for cluster: %s\n", run.ClusterName)
						}

						if appUri != "" {
							fmt.Printf("Set app URI: %s\n", appUri)
							run.SparkExtension.AppUri = &appUri
						}
					}
				}

				ew.setEMRMetricsUri(&run)
				fmt.Printf("Updating run %s\n", run.RunID)

				run, err = ew.sm.UpdateRun(run.RunID, run)
				if err != nil {
					fmt.Printf("Error updating run: %v\n", err)
					_ = ew.log.Log("message", "error saving kubernetes events", "emrJobId", emrJobId, "error", fmt.Sprintf("%+v", err))
				} else {
					fmt.Println("Successfully updated run")
				}

				if run.PodEvents != nil && len(*run.PodEvents) >= ew.emrMaxPodEvents {
					fmt.Printf("Reached max pod events (%d), terminating run\n", ew.emrMaxPodEvents)
					_ = ew.emrEngine.Terminate(run)
				}
			}
		} else {
			fmt.Println("No EMR job ID found, skipping run update")
		}

		fmt.Println("Completing event processing")
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
		fmt.Println("MetricsURI", &metricsUri)
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

func (ew *eventsWorker) processEvent(kubernetesEvent state.KubernetesEvent) {
	runId := kubernetesEvent.InvolvedObject.Labels.JobName
	if strings.HasPrefix(runId, "eks-spark") || len(runId) == 0 {
		ew.processEMRPodEvents(kubernetesEvent)
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
