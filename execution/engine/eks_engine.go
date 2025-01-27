package engine

import (
	"bytes"
	"context"
	"fmt"
	utils "github.com/stitchfix/flotilla-os/execution"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/clients/metrics"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/adapter"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
	kubernetestrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/k8s.io/client-go/kubernetes"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type EKSExecutionEngine struct {
	stateManager    state.Manager
	kClients        map[string]kubernetes.Clientset
	metricsClients  map[string]metricsv.Clientset
	adapter         adapter.EKSAdapter
	qm              queue.Manager
	log             flotillaLog.Logger
	jobQueue        string
	jobNamespace    string
	jobTtl          int
	jobSA           string
	jobARAEnabled   bool
	schedulerName   string
	serializer      *k8sJson.Serializer
	s3Client        *s3.S3
	s3Bucket        string
	s3BucketRootDir string
	statusQueue     string
}

func (ee *EKSExecutionEngine) Initialize(conf config.Config) error {
	ee.kClients = make(map[string]kubernetes.Clientset)
	ee.metricsClients = make(map[string]metricsv.Clientset)

	clusters, err := ee.stateManager.ListClusterStates()
	if err != nil {
		return errors.Wrap(err, "error listing clusters from state manager")
	}

	for _, cluster := range clusters {
		k8sConfig, err := utils.GetClusterConfig(cluster.Name, cluster.Region)

		if err != nil {
			return err
		}

		k8sConfig.WrapTransport = kubernetestrace.WrapRoundTripper
		kClient, err := kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			return err
		}

		ee.kClients[cluster.Name] = *kClient
		ee.metricsClients[cluster.Name] = *metricsv.NewForConfigOrDie(k8sConfig)

	}

	ee.jobQueue = conf.GetString("eks_job_queue")
	ee.schedulerName = "default-scheduler"

	if conf.IsSet("eks_scheduler_name") {
		ee.schedulerName = conf.GetString("eks_scheduler_name")
	}
	if conf.IsSet("eks_status_queue") {
		ee.statusQueue = conf.GetString("eks_status_queue")
	}
	ee.jobNamespace = conf.GetString("eks_job_namespace")
	ee.jobTtl = conf.GetInt("eks_job_ttl")
	ee.jobSA = conf.GetString("eks_default_service_account")
	ee.jobARAEnabled = true

	adapt, err := adapter.NewEKSAdapter()
	if err != nil {
		return err
	}

	ee.serializer = k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory, nil, nil,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	awsRegion := conf.GetString("eks_manifest_storage_options_region")
	awsConfig := &aws.Config{Region: aws.String(awsRegion)}
	sess := awstrace.WrapSession(session.Must(session.NewSessionWithOptions(
		session.Options{Config: *awsConfig})))
	ee.s3Client = s3.New(sess, aws.NewConfig().WithRegion(awsRegion))
	ee.s3Bucket = conf.GetString("eks_manifest_storage_options_s3_bucket_name")
	ee.s3BucketRootDir = conf.GetString("eks_manifest_storage_options_s3_bucket_root_dir")

	ee.adapter = adapt

	return nil
}

func (ee *EKSExecutionEngine) Execute(executable state.Executable, run state.Run, manager state.Manager) (state.Run, bool, error) {
	ee.log.Log("engine_received_tier", run.Tier)
	if run.Tier == "" {
		run.Tier = state.Tier4
	}

	clusters, err := ee.stateManager.ListClusterStates()
	if err != nil {
		return run, true, errors.Wrap(err, "error listing clusters")
	}
	selectedCluster := ee.selectCluster(clusters, run)
	if selectedCluster == nil {
		return run, true, errors.New("no eligible clusters found for run requirements")
	}
	fmt.Println("Selected cluster: ", selectedCluster.Name)
	run.ClusterName = selectedCluster.Name

	// Set namespace or EMR virtual cluster based on engine type
	if run.Engine != nil && *run.Engine == state.EKSSparkEngine {
		if selectedCluster.EMRVirtualCluster == "" {
			return run, true, errors.Errorf("cluster %s has no EMR virtual cluster configured", selectedCluster.Name)
		}
	} else {
		if selectedCluster.Namespace == "" {
			return run, true, errors.Errorf("cluster %s has no namespace configured", selectedCluster.Name)
		}
		run.Namespace = &selectedCluster.Namespace
	}

	ctx := context.Background()
	if run.ServiceAccount == nil {
		run.ServiceAccount = aws.String(ee.jobSA)
	}

	job, err := ee.adapter.AdaptFlotillaDefinitionAndRunToJob(executable, run, ee.schedulerName, manager, ee.jobARAEnabled)

	kClient, err := ee.getKClient(run)
	if err != nil {
		exitReason := fmt.Sprintf("Invalid cluster name - %s", run.ClusterName)
		run.ExitReason = &exitReason
		return run, false, err
	}

	result, err := kClient.BatchV1().Jobs(*run.Namespace).Create(ctx, &job, metav1.CreateOptions{})

	if err != nil {
		// Job is already submitted, don't retry
		if strings.Contains(strings.ToLower(err.Error()), "already exists") {
			return run, false, nil
		}

		// Job spec is invalid, don't retry
		if strings.Contains(strings.ToLower(err.Error()), "is invalid") {
			exitReason := err.Error()
			run.ExitReason = &exitReason
			return run, false, err
		}

		// Legitimate submit error, retryable.
		_ = metrics.Increment(metrics.EngineEKSExecute, []string{string(metrics.StatusFailure)}, 1)
		return run, true, err
	}

	var b0 bytes.Buffer
	err = ee.serializer.Encode(result, &b0)
	if err == nil {
		putObject := s3.PutObjectInput{
			Bucket:      aws.String(ee.s3Bucket),
			Body:        bytes.NewReader(b0.Bytes()),
			Key:         aws.String(fmt.Sprintf("%s/%s/%s.yaml", ee.s3BucketRootDir, run.RunID, run.RunID)),
			ContentType: aws.String("text/yaml"),
		}
		_, err = ee.s3Client.PutObject(&putObject)

		if err != nil {
			_ = ee.log.Log("s3_upload_error", "error", err.Error())
		}
	}
	_ = metrics.Increment(metrics.EngineEKSExecute, []string{string(metrics.StatusSuccess)}, 1)

	run, _ = ee.getPodName(run)
	// Set status to running.
	adaptedRun, err := ee.adapter.AdaptJobToFlotillaRun(result, run, nil)

	if err != nil {
		return adaptedRun, false, err
	}

	adaptedRun.Status = state.StatusRunning
	return adaptedRun, false, nil
}

func (ee *EKSExecutionEngine) selectCluster(clusters []state.ClusterMetadata, run state.Run) *state.ClusterMetadata {
	for _, cluster := range clusters {
		if !ee.isClusterEligible(cluster, run) {
			continue
		}
		return &cluster
	}
	return nil
}

func (ee *EKSExecutionEngine) isClusterEligible(cluster state.ClusterMetadata, run state.Run) bool {
	if cluster.Status != state.StatusActive {
		return false
	}

	if !containsTier(cluster.AllowedTiers, run.Tier) {
		return false
	}

	return true
}

func containsTier(tiers []state.Tier, tier state.Tier) bool {
	for _, t := range tiers {
		if t == tier {
			return true
		}
	}
	return false
}

func (ee *EKSExecutionEngine) getPodName(run state.Run) (state.Run, error) {
	podList, err := ee.getPodList(run)

	if err != nil {
		return run, err
	}

	if podList != nil && podList.Items != nil && len(podList.Items) > 0 {
		pod := podList.Items[len(podList.Items)-1]
		run.PodName = &pod.Name
		run.Namespace = &pod.Namespace
		if pod.Spec.Containers != nil && len(pod.Spec.Containers) > 0 {
			container := pod.Spec.Containers[len(pod.Spec.Containers)-1]
			cpu := container.Resources.Requests.Cpu().ScaledValue(resource.Milli)
			cpuLimit := container.Resources.Limits.Cpu().ScaledValue(resource.Milli)
			run.Cpu = &cpu
			run.CpuLimit = &cpuLimit
			run = ee.getInstanceDetails(pod, run)
			mem := container.Resources.Requests.Memory().ScaledValue(resource.Mega)
			run.Memory = &mem
			memLimit := container.Resources.Limits.Memory().ScaledValue(resource.Mega)
			run.MemoryLimit = &memLimit
		}
	}
	return run, nil
}

func (ee *EKSExecutionEngine) getInstanceDetails(pod v1.Pod, run state.Run) state.Run {
	if len(pod.Spec.NodeName) > 0 {
		run.InstanceDNSName = pod.Spec.NodeName
	}
	return run
}

func (ee *EKSExecutionEngine) getPodList(run state.Run) (*v1.PodList, error) {
	ctx := context.Background()
	kClient, err := ee.getKClient(run)
	if err != nil {
		return &v1.PodList{}, err
	}

	if run.PodName != nil {
		pod, err := kClient.CoreV1().Pods(*run.Namespace).Get(ctx, *run.PodName, metav1.GetOptions{})
		if pod != nil {
			return &v1.PodList{Items: []v1.Pod{*pod}}, err
		}
	} else {
		if run.QueuedAt == nil {
			return &v1.PodList{}, err
		}
		queuedAt := *run.QueuedAt
		if time.Now().After(queuedAt.Add(time.Minute * time.Duration(5))) {
			podList, err := kClient.CoreV1().Pods(*run.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("job-name=%s", run.RunID),
			})
			return podList, err
		}
	}
	return &v1.PodList{}, err
}

func (ee *EKSExecutionEngine) getKClient(run state.Run) (kubernetes.Clientset, error) {
	if run.ClusterName == "" {
		return kubernetes.Clientset{}, errors.New("Cluster name is empty")
	}
	kClient, ok := ee.kClients[run.ClusterName]
	if !ok {
		return kubernetes.Clientset{}, errors.New(fmt.Sprintf("Invalid cluster name - %s", run.ClusterName))
	}
	return kClient, nil
}

func (ee *EKSExecutionEngine) Terminate(run state.Run) error {
	ctx := context.Background()
	gracePeriod := int64(300)
	deletionPropagation := metav1.DeletePropagationBackground
	_ = ee.log.Log("terminating run=", run.RunID)
	deleteOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
		PropagationPolicy:  &deletionPropagation,
	}

	kClient, err := ee.getKClient(run)
	if err != nil {
		exitReason := fmt.Sprintf(err.Error())
		run.ExitReason = &exitReason
		return err
	}

	_ = kClient.BatchV1().Jobs(ee.jobNamespace).Delete(ctx, run.RunID, *deleteOptions)
	if run.PodName != nil {
		_ = kClient.CoreV1().Pods(ee.jobNamespace).Delete(ctx, *run.PodName, *deleteOptions)
	}

	_ = metrics.Increment(metrics.EngineEKSTerminate, []string{string(metrics.StatusSuccess)}, 1)
	return nil
}

func (ee *EKSExecutionEngine) Enqueue(run state.Run) error {
	// Get qurl
	qurl, err := ee.qm.QurlFor(ee.jobQueue, false)
	if err != nil {
		_ = metrics.Increment(metrics.EngineEKSEnqueue, []string{string(metrics.StatusFailure)}, 1)
		return errors.Wrapf(err, "problem getting queue url for [%s]", run.ClusterName)
	}

	// Queue run
	if err = ee.qm.Enqueue(qurl, run); err != nil {
		_ = metrics.Increment(metrics.EngineEKSEnqueue, []string{string(metrics.StatusFailure)}, 1)
		return errors.Wrapf(err, "problem enqueing run [%s] to queue [%s]", run.RunID, qurl)
	}

	_ = metrics.Increment(metrics.EngineEKSEnqueue, []string{string(metrics.StatusSuccess)}, 1)
	return nil
}

func (ee *EKSExecutionEngine) PollRuns() ([]RunReceipt, error) {
	qurl, err := ee.qm.QurlFor(ee.jobQueue, false)
	if err != nil {
		return nil, errors.Wrap(err, "problem listing queues to poll")
	}
	queues := []string{qurl}
	var runs []RunReceipt
	for _, qurl := range queues {
		//
		// Get new queued Run
		//
		runReceipt, err := ee.qm.ReceiveRun(qurl)

		if err != nil {
			return runs, errors.Wrapf(err, "problem receiving run from queue url [%s]", qurl)
		}

		if runReceipt.Run == nil {
			continue
		}

		runs = append(runs, RunReceipt{runReceipt})
	}
	return runs, nil
}

// PollStatus is a dummy function as EKS does not emit task status
// change events.
func (ee *EKSExecutionEngine) PollStatus() (RunReceipt, error) {
	return RunReceipt{}, nil
}

// Reads off SQS queue and generates a Run object based on the runId
func (ee *EKSExecutionEngine) PollRunStatus() (state.Run, error) {
	return state.Run{}, nil
}

// Define returns a blank task definition and an error for the EKS engine.
func (ee *EKSExecutionEngine) Define(td state.Definition) (state.Definition, error) {
	return td, errors.New("Definition of tasks are only for ECSs.")
}

// Deregister returns an error for the EKS engine.
func (ee *EKSExecutionEngine) Deregister(definition state.Definition) error {
	return errors.Errorf("EKSExecutionEngine does not allow for deregistering of task definitions.")
}

func (ee *EKSExecutionEngine) Get(run state.Run) (state.Run, error) {
	ctx := context.Background()
	kClient, err := ee.getKClient(run)
	if err != nil {
		return state.Run{}, err
	}
	job, err := kClient.BatchV1().Jobs(ee.jobNamespace).Get(ctx, run.RunID, metav1.GetOptions{})

	if err != nil {
		return state.Run{}, errors.Errorf("error getting kubernetes job %s", err)
	}
	updates, err := ee.adapter.AdaptJobToFlotillaRun(job, run, nil)

	if err != nil {
		return state.Run{}, errors.Errorf("error adapting kubernetes job to flotilla run %s", err)
	}

	return updates, nil
}

func (ee *EKSExecutionEngine) GetEvents(run state.Run) (state.PodEventList, error) {
	ctx := context.Background()
	if run.PodName == nil {
		return state.PodEventList{}, nil
	}
	kClient, err := ee.getKClient(run)
	if err != nil {
		return state.PodEventList{}, err
	}

	eventList, err := kClient.CoreV1().Events(ee.jobNamespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("involvedObject.name==%s", *run.PodName)})
	if err != nil {
		return state.PodEventList{}, errors.Errorf("error getting kubernetes event for flotilla run %s", err)
	}

	var podEvents []state.PodEvent
	for _, e := range eventList.Items {
		eTime := e.FirstTimestamp.Time
		runEvent := state.PodEvent{
			Message:      e.Message,
			Timestamp:    &eTime,
			EventType:    e.Type,
			Reason:       e.Reason,
			SourceObject: e.ObjectMeta.Name,
		}

		if strings.Contains(e.Reason, "TriggeredScaleUp") {
			source := fmt.Sprintf("source:%s", e.ObjectMeta.Name)
			_ = metrics.Increment(metrics.EngineEKSNodeTriggeredScaledUp, []string{source}, 1)
		}
		podEvents = append(podEvents, runEvent)
	}

	podEventList := state.PodEventList{
		Total:     len(podEvents),
		PodEvents: podEvents,
	}

	return podEventList, nil
}

func (ee *EKSExecutionEngine) FetchPodMetrics(run state.Run) (state.Run, error) {
	ctx := context.Background()
	if run.PodName != nil {
		metricsClient, ok := ee.metricsClients[run.ClusterName]
		if !ok {
			return run, errors.New("Metrics client not defined.")
		}
		start := time.Now()
		podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(ee.jobNamespace).Get(ctx, *run.PodName, metav1.GetOptions{})
		_ = metrics.Timing(metrics.StatusWorkerFetchMetrics, time.Since(start), []string{run.ClusterName}, 1)

		if err != nil {
			return run, err
		}
		if len(podMetrics.Containers) > 0 {
			containerMetrics := podMetrics.Containers[0]
			mem := containerMetrics.Usage.Memory().ScaledValue(resource.Mega)
			if run.MaxMemoryUsed == nil || *run.MaxMemoryUsed == 0 || *run.MaxMemoryUsed < mem {
				run.MaxMemoryUsed = &mem
			}

			cpu := containerMetrics.Usage.Cpu().MilliValue()
			if run.MaxCpuUsed == nil || *run.MaxCpuUsed == 0 || *run.MaxCpuUsed < cpu {
				run.MaxCpuUsed = &cpu
			}
		}
		return run, nil
	}
	return run, errors.New("no pod associated with the run.")
}

func (ee *EKSExecutionEngine) FetchUpdateStatus(run state.Run) (state.Run, error) {
	ctx := context.Background()
	kClient, err := ee.getKClient(run)
	if err != nil {
		return state.Run{}, err
	}

	start := time.Now()
	job, err := kClient.BatchV1().Jobs(ee.jobNamespace).Get(ctx, run.RunID, metav1.GetOptions{})
	_ = metrics.Timing(metrics.StatusWorkerGetJob, time.Since(start), []string{run.ClusterName}, 1)

	if err != nil {
		return run, err
	}

	var mostRecentPod *v1.Pod
	var mostRecentPodCreationTimestamp metav1.Time

	start = time.Now()
	podList, err := ee.getPodList(run)
	_ = metrics.Timing(metrics.StatusWorkerGetPodList, time.Since(start), []string{run.ClusterName}, 1)

	if err == nil && podList != nil && podList.Items != nil && len(podList.Items) > 0 {
		// Iterate over associated pods to find the most recent.
		for _, p := range podList.Items {
			if mostRecentPodCreationTimestamp.Before(&p.CreationTimestamp) || len(podList.Items) == 1 {
				mostRecentPod = &p
				mostRecentPodCreationTimestamp = p.CreationTimestamp
			}
		}

		// If the run doesn't have an associated pod name yet OR
		// there is a newer pod (i.e. the old pod was killed),
		// update it.
		if mostRecentPod != nil && (run.PodName == nil || mostRecentPod.Name != *run.PodName) {
			if run.PodName != nil && mostRecentPod.Name != *run.PodName {
				_ = metrics.Increment(metrics.EngineEKSRunPodnameChange, []string{}, 1)
			}

			run.PodName = &mostRecentPod.Name
			run = ee.getInstanceDetails(*mostRecentPod, run)
		}

		// Pod didn't change, but Instance information is not populated.
		if mostRecentPod != nil && len(run.InstanceDNSName) == 0 {
			run = ee.getInstanceDetails(*mostRecentPod, run)
		}

		if mostRecentPod != nil && mostRecentPod.Spec.Containers != nil && len(mostRecentPod.Spec.Containers) > 0 {
			container := mostRecentPod.Spec.Containers[len(mostRecentPod.Spec.Containers)-1]
			cpu := container.Resources.Requests.Cpu().ScaledValue(resource.Milli)
			run.Cpu = &cpu
			mem := container.Resources.Requests.Memory().ScaledValue(resource.Mega)
			run.Memory = &mem
			cpuLimit := container.Resources.Limits.Cpu().ScaledValue(resource.Milli)
			run.CpuLimit = &cpuLimit
			memLimit := container.Resources.Limits.Memory().ScaledValue(resource.Mega)
			run.MemoryLimit = &memLimit
		}
	}

	//run, _ = ee.FetchPodMetrics(run)
	hoursBack := time.Now().Add(-24 * time.Hour)

	start = time.Now()
	var events state.PodEventList
	//events, err = ee.GetEvents(run)
	_ = metrics.Timing(metrics.StatusWorkerGetEvents, time.Since(start), []string{run.ClusterName}, 1)

	if err == nil && len(events.PodEvents) > 0 {
		newEvents := events.PodEvents
		if run.PodEvents != nil && len(*run.PodEvents) > 0 {
			priorEvents := *run.PodEvents
			for _, newEvent := range newEvents {
				unseen := true
				for _, priorEvent := range priorEvents {
					if priorEvent.Equal(newEvent) {
						unseen = false
						break
					}
				}
				if unseen {
					priorEvents = append(priorEvents, newEvent)
				}
			}
			run.PodEvents = &priorEvents
		} else {
			run.PodEvents = &newEvents
		}
	}

	if run.PodEvents != nil {
		attemptCount := int64(0)
		for _, podEvent := range *run.PodEvents {
			if strings.Contains(podEvent.Reason, "Scheduled") {
				attemptCount = attemptCount + 1
			}
		}
		run.AttemptCount = &attemptCount
	}

	// Handle edge case for dangling jobs.
	// Run used to have a pod and now it is not there, job is older than 24 hours. Terminate it.
	if err == nil && podList != nil && podList.Items != nil && len(podList.Items) == 0 && run.PodName != nil && run.QueuedAt.Before(hoursBack) {
		err = ee.Terminate(run)
		if err == nil {
			job.Status.Failed = 1
			mostRecentPod = nil
		}
	}

	return ee.adapter.AdaptJobToFlotillaRun(job, run, mostRecentPod)
}
