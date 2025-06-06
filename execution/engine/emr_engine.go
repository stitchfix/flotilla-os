package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/clients/metrics"
	"github.com/stitchfix/flotilla-os/utils"

	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	awstrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go/aws"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/kubernetes/scheme"
	"regexp"
	"strings"
)

// EMRExecutionEngine submits runs to EMR-EKS.
type EMRExecutionEngine struct {
	sqsQueueManager     queue.Manager
	log                 flotillaLog.Logger
	emrJobQueue         string
	emrJobNamespace     string
	emrJobRoleArn       map[string]string
	emrJobSA            string
	emrVirtualClusters  map[string]string
	emrContainersClient *emrcontainers.EMRContainers
	schedulerName       string
	s3Client            *s3.S3
	awsRegion           string
	s3LogsBucket        string
	s3EventLogPath      string
	s3LogsBasePath      string
	s3ManifestBucket    string
	s3ManifestBasePath  string
	serializer          *k8sJson.Serializer
	clusters            []string
	driverInstanceType  string
	kClients            map[string]kubernetes.Clientset
	clusterManager      *DynamicClusterManager
	stateManager        state.Manager
	redisClient         *redis.Client
}

// Initialize configures the EMRExecutionEngine and initializes internal clients
func (emr *EMRExecutionEngine) Initialize(conf config.Config) error {

	emr.emrVirtualClusters = make(map[string]string)
	emr.emrVirtualClusters = conf.GetStringMapString("emr_virtual_clusters")

	emr.emrJobQueue = conf.GetString("emr_job_queue")
	emr.emrJobNamespace = conf.GetString("emr_job_namespace")
	emr.emrJobRoleArn = conf.GetStringMapString("emr_job_role_arn")
	emr.awsRegion = conf.GetString("emr_aws_region")
	emr.s3LogsBucket = conf.GetString("emr_log_bucket")
	emr.s3LogsBasePath = conf.GetString("emr_log_base_path")
	emr.s3EventLogPath = conf.GetString("emr_log_event_log_path")
	emr.s3ManifestBucket = conf.GetString("emr_manifest_bucket")
	emr.s3ManifestBasePath = conf.GetString("emr_manifest_base_path")
	emr.emrJobSA = conf.GetString("emr_default_service_account")
	emr.schedulerName = conf.GetString("eks_scheduler_name")
	emr.driverInstanceType = conf.GetString("emr_driver_instance_type")
	awsConfig := &aws.Config{Region: aws.String(emr.awsRegion)}
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: *awsConfig}))
	sess = awstrace.WrapSession(sess)
	emr.s3Client = s3.New(sess, aws.NewConfig().WithRegion(emr.awsRegion))
	emr.emrContainersClient = emrcontainers.New(sess, aws.NewConfig().WithRegion(emr.awsRegion))

	emr.serializer = k8sJson.NewSerializerWithOptions(
		k8sJson.SimpleMetaFactory{}, nil, nil,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)

	clusterManager, err := NewDynamicClusterManager(
		emr.awsRegion,
		emr.log,
		emr.stateManager,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create dynamic cluster manager")
	}
	emr.clusterManager = clusterManager

	// Get static clusters if configured
	var staticClusters []string
	if conf.IsSet("eks_clusters") {
		clusters := strings.Split(conf.GetString("eks_clusters"), ",")
		for i := range clusters {
			staticClusters = append(staticClusters, strings.TrimSpace(clusters[i]))
		}
	}

	// Initialize all clusters (both static and dynamic)
	if err := clusterManager.InitializeClusters(context.Background(), staticClusters); err != nil {
		emr.log.Log("message", "failed to initialize clusters", "error", err.Error())
	}

	return nil
}

func (emr *EMRExecutionEngine) getKClient(run state.Run) (kubernetes.Clientset, error) {
	kClient, err := emr.clusterManager.GetKubernetesClient(run.ClusterName)
	if err != nil {
		return kubernetes.Clientset{}, errors.Wrapf(err, "failed to get Kubernetes client for cluster %s", run.ClusterName)
	}
	return kClient, nil
}
func (emr *EMRExecutionEngine) Execute(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) (state.Run, bool, error) {
	var span tracer.Span
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span = utils.TraceJob(ctx, "flotilla.job.emr_execute", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	run = emr.estimateExecutorCount(run, manager)
	run = emr.estimateMemoryResources(ctx, run, manager)

	if run.ServiceAccount == nil || *run.ServiceAccount == "" {
		run.ServiceAccount = aws.String(emr.emrJobSA)
	}

	if run.CommandHash != nil && run.NodeLifecycle != nil && *run.NodeLifecycle == state.SpotLifecycle {
		nodeType, err := manager.GetNodeLifecycle(ctx, run.DefinitionID, *run.CommandHash)
		if err == nil && nodeType == state.OndemandLifecycle {
			run.NodeLifecycle = &state.OndemandLifecycle
		}
	}

	startJobRunInput, err := emr.generateEMRStartJobRunInput(ctx, executable, run, manager)
	emrJobManifest := aws.String(fmt.Sprintf("%s/%s/%s.json", emr.s3ManifestBasePath, run.RunID, "start-job-run-input"))
	obj, err := json.MarshalIndent(startJobRunInput, "", "\t")
	if err == nil {
		emrJobManifest = emr.writeStringToS3(emrJobManifest, obj)
	}

	emr.log.Log("message", "Start EMR JobRun", "ExecutionRoleArn", startJobRunInput.ExecutionRoleArn)
	tierTag := fmt.Sprintf("tier:%s", run.Tier)

	startJobRunOutput, err := emr.emrContainersClient.StartJobRun(&startJobRunInput)
	if err == nil {
		run.SparkExtension.VirtualClusterId = startJobRunOutput.VirtualClusterId
		run.SparkExtension.EMRJobId = startJobRunOutput.Id
		run.SparkExtension.EMRJobManifest = emrJobManifest
		run.Status = state.StatusQueued
		_ = metrics.Increment(metrics.EngineEMRExecute, []string{string(metrics.StatusSuccess), tierTag}, 1)
	} else {
		run.ExitReason = aws.String(fmt.Sprintf("%v", err))
		run.ExitCode = aws.Int64(-1)
		run.StartedAt = run.QueuedAt
		run.FinishedAt = run.QueuedAt
		run.Status = state.StatusStopped
		_ = emr.log.Log("EMR job submission error", "error", err.Error())
		_ = metrics.Increment(metrics.EngineEKSExecute, []string{string(metrics.StatusFailure), tierTag}, 1)
		return run, false, err
	}
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	} else {
		span.SetTag("emr.job_id", *run.SparkExtension.EMRJobId)
		span.SetTag("emr.virtual_cluster_id", *run.SparkExtension.VirtualClusterId)
		utils.TagJobRun(span, run)
	}
	return run, false, nil
}

func (emr *EMRExecutionEngine) generateApplicationConf(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) []*emrcontainers.Configuration {
	if ctx == nil {
		ctx = context.Background()
	}
	sparkDefaults := map[string]*string{
		"spark.kubernetes.driver.podTemplateFile":   emr.driverPodTemplate(ctx, executable, run, manager),
		"spark.kubernetes.executor.podTemplateFile": emr.executorPodTemplate(ctx, executable, run, manager),
		"spark.kubernetes.container.image":          &run.Image,
		"spark.eventLog.dir":                        aws.String(fmt.Sprintf("s3://%s/%s", emr.s3LogsBucket, emr.s3EventLogPath)),
		"spark.history.fs.logDirectory":             aws.String(fmt.Sprintf("s3://%s/%s", emr.s3LogsBucket, emr.s3EventLogPath)),
		"spark.eventLog.enabled":                    aws.String("true"),
		"spark.default.parallelism":                 aws.String("256"),
		"spark.sql.shuffle.partitions":              aws.String("256"),

		// PrometheusServlet metrics config
		"spark.metrics.conf.*.sink.prometheusServlet.class": aws.String("org.apache.spark.metrics.sink.PrometheusServlet"),
		"spark.metrics.conf.*.sink.prometheusServlet.path":  aws.String("/metrics/driver/prometheus"),
		"master.sink.prometheusServlet.path":                aws.String("/metrics/master/prometheus"),
		"applications.sink.prometheusServlet.path":          aws.String("/metrics/applications/prometheus"),

		// Metrics grouped per component instance and source namespace e.g., Component instance = Driver or Component instance = shuffleService
		"spark.kubernetes.driver.service.annotation.prometheus.io/port":   aws.String("4040"),
		"spark.kubernetes.driver.service.annotation.prometheus.io/path":   aws.String("/metrics/driver/prometheus/"),
		"spark.kubernetes.driver.service.annotation.prometheus.io/scrape": aws.String("true"),

		// Executor-level metrics are sent from each executor to the driver. Prometheus endpoint at: /metrics/executors/prometheus
		"spark.kubernetes.driver.annotation.prometheus.io/scrape": aws.String("true"),
		"spark.kubernetes.driver.annotation.prometheus.io/path":   aws.String("/metrics/executors/prometheus/"),
		"spark.kubernetes.driver.annotation.prometheus.io/port":   aws.String("4040"),
		"spark.ui.prometheus.enabled":                             aws.String("true"),
	}

	hiveDefaults := map[string]*string{}

	for _, k := range run.SparkExtension.ApplicationConf {
		sparkDefaults[*k.Name] = k.Value
	}
	if run.SparkExtension.HiveConf != nil {
		for _, k := range run.SparkExtension.HiveConf {
			if k.Name != nil && k.Value != nil {
				hiveDefaults[*k.Name] = k.Value
			}
		}
	}

	return []*emrcontainers.Configuration{
		{
			Classification: aws.String("spark-defaults"),
			Properties:     sparkDefaults,
		},
		{
			Classification: aws.String("spark-hive-site"),
			Properties:     hiveDefaults,
		},
	}
}

func (emr *EMRExecutionEngine) generateEMRStartJobRunInput(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) (emrcontainers.StartJobRunInput, error) {
	roleArn := emr.emrJobRoleArn[*run.ServiceAccount]
	if ctx == nil {
		ctx = context.Background()
	}
	dbClusters, err := emr.stateManager.ListClusterStates(ctx)
	if err != nil {
		emr.log.Log("message", "failed to get clusters from database", "error", err.Error())
		return emrcontainers.StartJobRunInput{}, err
	}
	var clusterID string
	clusterFound := false
	for _, cluster := range dbClusters {
		if cluster.Namespace == emr.emrJobNamespace && cluster.Name == run.ClusterName {
			clusterID = cluster.EMRVirtualCluster
			if cluster.SparkServerURI != "" {
				run.SparkExtension.SparkServerURI = aws.String(cluster.SparkServerURI)
			}
			clusterFound = true
			break
		}
	}
	if !clusterFound {
		clusterID = emr.emrVirtualClusters[run.ClusterName]
	}

	if clusterID == "" {
		return emrcontainers.StartJobRunInput{}, fmt.Errorf("EMR virtual cluster ID not found for EKS cluster: %s", run.ClusterName)
	}
	startJobRunInput := emrcontainers.StartJobRunInput{
		ClientToken: &run.RunID,
		ConfigurationOverrides: &emrcontainers.ConfigurationOverrides{
			MonitoringConfiguration: &emrcontainers.MonitoringConfiguration{
				PersistentAppUI: aws.String(emrcontainers.PersistentAppUIEnabled),
				S3MonitoringConfiguration: &emrcontainers.S3MonitoringConfiguration{
					LogUri: aws.String(fmt.Sprintf("s3://%s/%s", emr.s3LogsBucket, emr.s3LogsBasePath)),
				},
			},
			ApplicationConfiguration: emr.generateApplicationConf(ctx, executable, run, manager),
		},
		ExecutionRoleArn: &roleArn,
		JobDriver: &emrcontainers.JobDriver{
			SparkSubmitJobDriver: &emrcontainers.SparkSubmitJobDriver{
				EntryPoint:            run.SparkExtension.SparkSubmitJobDriver.EntryPoint,
				EntryPointArguments:   run.SparkExtension.SparkSubmitJobDriver.EntryPointArguments,
				SparkSubmitParameters: emr.sparkSubmitParams(run),
			}},
		Name:             &run.RunID,
		ReleaseLabel:     run.SparkExtension.EMRReleaseLabel,
		VirtualClusterId: &clusterID,
	}
	return startJobRunInput, nil
}

func (emr *EMRExecutionEngine) generateTags(run state.Run) map[string]*string {
	tags := make(map[string]*string)
	if run.Env != nil && len(*run.Env) > 0 {
		for _, ev := range *run.Env {
			name := emr.sanitizeEnvVar(ev.Name)
			space := regexp.MustCompile(`\s+`)
			if len(ev.Value) < 256 && len(name) < 128 {
				tags[name] = aws.String(space.ReplaceAllString(ev.Value, ""))
			}
		}
	}
	return tags
}

// generates volumes and volumemounts depending on cluster name.
// TODO cleanup after migration
func generateVolumesForCluster(clusterName string, isEmptyDir bool) ([]v1.Volume, []v1.VolumeMount) {
	var volumes []v1.Volume
	var volumeMounts []v1.VolumeMount

	if isEmptyDir {
		// Use a emptyDir volume
		specificVolume := v1.Volume{
			Name: "shared-lib-volume",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &(v1.EmptyDirVolumeSource{}),
			},
		}

		volumes = append(volumes, specificVolume)
	} else {
		// Use the persistent volume claim
		sharedLibVolume := v1.Volume{
			Name: "shared-lib-volume",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: "s3-claim",
				},
			},
		}
		volumes = append(volumes, sharedLibVolume)
	}
	volumeMount := v1.VolumeMount{
		Name:      "shared-lib-volume",
		MountPath: "/var/lib/app",
	}
	volumeMounts = append(volumeMounts, volumeMount)

	return volumes, volumeMounts
}

func (emr *EMRExecutionEngine) driverPodTemplate(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) *string {
	if ctx == nil {
		ctx = context.Background()
	}
	// Override driver pods to always be on ondemand nodetypes.
	run.NodeLifecycle = &state.OndemandLifecycle
	workingDir := "/var/lib/app"
	if run.SparkExtension != nil && run.SparkExtension.SparkSubmitJobDriver != nil && run.SparkExtension.SparkSubmitJobDriver.WorkingDir != nil {
		workingDir = *run.SparkExtension.SparkSubmitJobDriver.WorkingDir
	}

	volumes, volumeMounts := generateVolumesForCluster(run.ClusterName, true)

	podSpec := v1.PodSpec{
		TerminationGracePeriodSeconds: aws.Int64(90),
		Volumes:                       volumes,
		SchedulerName:                 emr.schedulerName,
		Containers: []v1.Container{
			{
				Name:         "spark-kubernetes-driver",
				Env:          emr.envOverrides(executable, run),
				VolumeMounts: volumeMounts,
				WorkingDir:   workingDir,
			},
		},
		InitContainers: []v1.Container{{
			Name:         fmt.Sprintf("init-driver-%s", run.RunID),
			Image:        run.Image,
			Env:          emr.envOverrides(executable, run),
			VolumeMounts: volumeMounts,
			Command:      emr.constructCmdSlice(run.SparkExtension.DriverInitCommand),
		}},
		RestartPolicy: v1.RestartPolicyNever,
		Affinity:      emr.constructAffinity(ctx, executable, run, manager, true),
		Tolerations:   emr.constructTolerations(executable, run),
	}

	if emr.driverInstanceType != "" {
		podSpec.NodeSelector = map[string]string{
			"node.kubernetes.io/instance-type": emr.driverInstanceType,
		}
	}

	labels := state.GetLabels(run)
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"karpenter.sh/do-not-evict": "true",
				"flotilla-run-id":           run.RunID,
			},
			Labels: labels,
		},
		Spec: podSpec,
	}

	key := aws.String(fmt.Sprintf("%s/%s/%s.yaml", emr.s3ManifestBasePath, run.RunID, "driver-template"))
	return emr.writeK8ObjToS3(&pod, key)
}

func (emr *EMRExecutionEngine) executorPodTemplate(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) *string {
	if ctx == nil {
		ctx = context.Background()
	}
	workingDir := "/var/lib/app"
	if run.SparkExtension != nil && run.SparkExtension.SparkSubmitJobDriver != nil && run.SparkExtension.SparkSubmitJobDriver.WorkingDir != nil {
		workingDir = *run.SparkExtension.SparkSubmitJobDriver.WorkingDir
	}

	labels := state.GetLabels(run)

	// TODO Remove after migration
	volumes, volumeMounts := generateVolumesForCluster(run.ClusterName, true)

	pod := v1.Pod{
		Status: v1.PodStatus{},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"karpenter.sh/do-not-evict": "true",
				"flotilla-run-id":           run.RunID},
			Labels: labels,
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: aws.Int64(90),
			Volumes:                       volumes,
			SchedulerName:                 emr.schedulerName,
			Containers: []v1.Container{
				{
					Name:         "spark-kubernetes-executor",
					Env:          emr.envOverrides(executable, run),
					VolumeMounts: volumeMounts,
					WorkingDir:   workingDir,
				},
			},
			InitContainers: []v1.Container{{
				Name:         fmt.Sprintf("init-executor-%s", run.RunID),
				Image:        run.Image,
				Env:          emr.envOverrides(executable, run),
				VolumeMounts: volumeMounts,
				Command:      emr.constructCmdSlice(run.SparkExtension.ExecutorInitCommand),
			}},
			RestartPolicy: v1.RestartPolicyNever,
			Affinity:      emr.constructAffinity(ctx, executable, run, manager, false),
			Tolerations:   emr.constructTolerations(executable, run),
		},
	}

	key := aws.String(fmt.Sprintf("%s/%s/%s.yaml", emr.s3ManifestBasePath, run.RunID, "executor-template"))
	return emr.writeK8ObjToS3(&pod, key)
}

func (emr *EMRExecutionEngine) writeK8ObjToS3(obj runtime.Object, key *string) *string {
	var b0 bytes.Buffer
	err := emr.serializer.Encode(obj, &b0)
	payload := bytes.ReplaceAll(b0.Bytes(), []byte("status: {}"), []byte(""))
	payload = bytes.ReplaceAll(payload, []byte("creationTimestamp: null"), []byte(""))
	payload = bytes.ReplaceAll(payload, []byte("resources: {}"), []byte(""))

	if err == nil {
		putObject := s3.PutObjectInput{
			Bucket:      aws.String(emr.s3ManifestBucket),
			Body:        bytes.NewReader(payload),
			Key:         key,
			ContentType: aws.String("text/yaml"),
		}
		_, err = emr.s3Client.PutObject(&putObject)
		if err != nil {
			_ = emr.log.Log("s3_upload_error", "error", err.Error())
		}
	}

	return aws.String(fmt.Sprintf("s3://%s/%s", emr.s3ManifestBucket, *key))
}

func (emr *EMRExecutionEngine) writeStringToS3(key *string, body []byte) *string {
	if body != nil && key != nil {
		putObject := s3.PutObjectInput{
			Bucket:      aws.String(emr.s3ManifestBucket),
			Body:        bytes.NewReader(body),
			Key:         key,
			ContentType: aws.String("text/yaml"),
		}
		_, err := emr.s3Client.PutObject(&putObject)
		if err != nil {
			_ = emr.log.Log("s3_upload_error", "error", err.Error())
		}
	}
	return aws.String(fmt.Sprintf("s3://%s/%s", emr.s3ManifestBucket, *key))
}

func (emr *EMRExecutionEngine) constructEviction(ctx context.Context, run state.Run, manager state.Manager) string {
	if ctx == nil {
		ctx = context.Background()
	}
	if run.NodeLifecycle != nil && *run.NodeLifecycle == state.OndemandLifecycle {
		return "false"
	}
	if run.CommandHash != nil {
		nodeType, err := manager.GetNodeLifecycle(ctx, run.DefinitionID, *run.CommandHash)
		if err == nil && nodeType == state.OndemandLifecycle {
			return "false"
		}
	}
	return "true"
}

func (emr *EMRExecutionEngine) constructTolerations(executable state.Executable, run state.Run) []v1.Toleration {
	tolerations := []v1.Toleration{}

	tolerations = append(tolerations, v1.Toleration{
		Key:      "emr",
		Operator: "Equal",
		Value:    "true",
		Effect:   "NoSchedule",
	})

	return tolerations
}

func (emr *EMRExecutionEngine) constructAffinity(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager, driver bool) *v1.Affinity {
	affinity := &v1.Affinity{}
	if ctx == nil {
		ctx = context.Background()
	}
	var requiredMatch []v1.NodeSelectorRequirement
	//todo move to config
	nodeLifecycleKey := "karpenter.sh/capacity-type"
	nodeArchKey := "kubernetes.io/arch"

	newCluster := true

	arch := []string{"amd64"}
	if run.Arch != nil && *run.Arch == "arm64" {
		arch = []string{"arm64"}
	}

	var nodeLifecycle []string
	nodePreference := "spot"
	if (run.NodeLifecycle != nil && *run.NodeLifecycle == state.OndemandLifecycle) || driver {
		nodeLifecycle = append(nodeLifecycle, "on-demand")
		nodePreference = "on-demand"
	} else {
		nodeLifecycle = append(nodeLifecycle, "spot", "on-demand")
	}

	if run.CommandHash != nil {
		nodeType, err := manager.GetNodeLifecycle(ctx, run.DefinitionID, *run.CommandHash)
		if err == nil && nodeType == state.OndemandLifecycle {
			nodeLifecycle = []string{"on-demand"}
		}
	}

	requiredMatch = append(requiredMatch, v1.NodeSelectorRequirement{
		Key:      nodeLifecycleKey,
		Operator: v1.NodeSelectorOpIn,
		Values:   nodeLifecycle,
	})

	requiredMatch = append(requiredMatch, v1.NodeSelectorRequirement{
		Key:      nodeArchKey,
		Operator: v1.NodeSelectorOpIn,
		Values:   arch,
	})

	//todo remove conditional after migration
	if newCluster {
		requiredMatch = append(requiredMatch, v1.NodeSelectorRequirement{
			Key:      "emr",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"true"},
		})
	}

	affinity = &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: requiredMatch,
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{{
				Weight: 50,
				Preference: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{{
						Key:      nodeLifecycleKey,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{nodePreference},
					}},
				},
			}},
		},
		PodAffinity: &v1.PodAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
				{
					Weight: 40,
					PodAffinityTerm: v1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"flotilla-run-id": run.RunID},
						},
						TopologyKey: "topology.kubernetes.io/zone",
					},
				},
			},
		},
	}
	return affinity
}

func (emr *EMRExecutionEngine) estimateExecutorCount(run state.Run, manager state.Manager) state.Run {
	return run
}

func setResourceSuffix(value string) string {
	if strings.Contains(value, "g") || strings.Contains(value, "m") {
		return strings.ToUpper(value)
	}
	if strings.Contains(value, "K") {
		return strings.ToLower(value)
	}
	return value
}

func (emr *EMRExecutionEngine) estimateMemoryResources(ctx context.Context, run state.Run, manager state.Manager) state.Run {
	if run.CommandHash == nil {
		return run
	}
	if ctx == nil {
		ctx = context.Background()
	}
	executorOOM, _ := manager.ExecutorOOM(ctx, run.DefinitionID, *run.CommandHash)
	driverOOM, _ := manager.DriverOOM(ctx, run.DefinitionID, *run.CommandHash)

	var sparkSubmitConf []state.Conf
	for _, k := range run.SparkExtension.SparkSubmitJobDriver.SparkSubmitConf {
		if *k.Name == "spark.executor.memory" && k.Value != nil {
			// 1.25x executor memory - OOM in the last 30 days
			if executorOOM {
				quantity := resource.MustParse(setResourceSuffix(*k.Value))
				quantity.Set(int64(float64(quantity.Value()) * 1.25))
				k.Value = aws.String(strings.ToLower(quantity.String()))
			} else {
				quantity := resource.MustParse(setResourceSuffix(*k.Value))
				minVal := resource.MustParse("1G")
				if quantity.MilliValue() > minVal.MilliValue() {
					quantity.Set(int64(float64(quantity.Value()) * 1.0))
					k.Value = aws.String(strings.ToLower(quantity.String()))
				}
			}
		}
		if driverOOM {
			//Bump up driver by 3x, jvm memory strings
			if *k.Name == "spark.driver.memory" && k.Value != nil {
				quantity := resource.MustParse(setResourceSuffix(*k.Value))
				quantity.Set(quantity.Value() * 3)
				k.Value = aws.String(strings.ToLower(quantity.String()))
			}
		}
		sparkSubmitConf = append(sparkSubmitConf, state.Conf{Name: k.Name, Value: k.Value})
	}
	run.SparkExtension.SparkSubmitJobDriver.SparkSubmitConf = sparkSubmitConf
	return run
}

func (emr *EMRExecutionEngine) sparkSubmitParams(run state.Run) *string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(" --name %s", run.RunID))

	for _, k := range run.SparkExtension.SparkSubmitJobDriver.SparkSubmitConf {
		buffer.WriteString(fmt.Sprintf(" --conf %s=%s", *k.Name, *k.Value))
	}

	buffer.WriteString(fmt.Sprintf(" --conf %s=%s", "spark.kubernetes.executor.podNamePrefix", run.RunID))
	buffer.WriteString(fmt.Sprintf(" --conf spark.log4j.rootLogger=DEBUG"))
	buffer.WriteString(fmt.Sprintf(" --conf spark.log4j.rootCategory=DEBUG"))

	if run.SparkExtension.SparkSubmitJobDriver.Class != nil {
		buffer.WriteString(fmt.Sprintf(" --class %s", *run.SparkExtension.SparkSubmitJobDriver.Class))
	}

	if len(run.SparkExtension.SparkSubmitJobDriver.Files) > 0 {
		files := strings.Join(run.SparkExtension.SparkSubmitJobDriver.Files, ",")
		buffer.WriteString(fmt.Sprintf(" --files %s", files))
	}

	if len(run.SparkExtension.SparkSubmitJobDriver.PyFiles) > 0 {
		files := strings.Join(run.SparkExtension.SparkSubmitJobDriver.PyFiles, ",")
		buffer.WriteString(fmt.Sprintf(" --py-files %s", files))
	}

	if len(run.SparkExtension.SparkSubmitJobDriver.Jars) > 0 {
		jars := strings.Join(run.SparkExtension.SparkSubmitJobDriver.Jars, ",")
		buffer.WriteString(fmt.Sprintf(" --jars %s", jars))
	}

	return aws.String(buffer.String())
}

func (emr *EMRExecutionEngine) Terminate(ctx context.Context, run state.Run) error {
	var span tracer.Span
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span = utils.TraceJob(ctx, "flotilla.job.emr_terminate", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	if run.Status == state.StatusStopped {
		return errors.New("Run is already in a stopped state.")
	}

	cancelJobRunInput := emrcontainers.CancelJobRunInput{
		Id:               run.SparkExtension.EMRJobId,
		VirtualClusterId: run.SparkExtension.VirtualClusterId,
	}
	tierTag := fmt.Sprintf("tier:%s", run.Tier)

	key := aws.String(fmt.Sprintf("%s/%s/%s.json", emr.s3ManifestBasePath, run.RunID, "cancel-job-run-input"))
	obj, err := json.Marshal(cancelJobRunInput)
	if err == nil {
		emr.writeStringToS3(key, obj)
	}

	_, err = emr.emrContainersClient.CancelJobRun(&cancelJobRunInput)
	if err != nil {
		_ = metrics.Increment(metrics.EngineEMRTerminate, []string{string(metrics.StatusFailure), tierTag}, 1)
		_ = emr.log.Log("EMR job termination error", "error", err.Error())
	}
	_ = metrics.Increment(metrics.EngineEMRTerminate, []string{string(metrics.StatusSuccess), tierTag}, 1)

	return err
}

func (emr *EMRExecutionEngine) Enqueue(ctx context.Context, run state.Run) error {
	var span tracer.Span
	ctx, span = utils.TraceJob(ctx, "flotilla.job.emr_enqueue", "")
	defer span.Finish()
	span.SetTag("job.run_id", run.RunID)
	span.SetTag("job.tier", run.Tier)
	utils.TagJobRun(span, run)
	tierTag := fmt.Sprintf("tier:%s", run.Tier)
	qurl, err := emr.sqsQueueManager.QurlFor(emr.emrJobQueue, false)
	if err != nil {
		_ = metrics.Increment(metrics.EngineEMREnqueue, []string{string(metrics.StatusFailure), tierTag}, 1)
		_ = emr.log.Log("EMR job enqueue error", "error", err.Error())
		return errors.Wrapf(err, "problem getting queue url for [%s]", run.ClusterName)
	}

	// Queue run
	if err = emr.sqsQueueManager.Enqueue(ctx, qurl, run); err != nil {
		_ = metrics.Increment(metrics.EngineEMREnqueue, []string{string(metrics.StatusFailure), tierTag}, 1)
		_ = emr.log.Log("EMR job enqueue error", "error", err.Error())
		return errors.Wrapf(err, "problem enqueing run [%s] to queue [%s]", run.RunID, qurl)
	}

	_ = metrics.Increment(metrics.EngineEMREnqueue, []string{string(metrics.StatusSuccess), tierTag}, 1)
	return nil
}

func (emr *EMRExecutionEngine) PollRuns(ctx context.Context) ([]RunReceipt, error) {
	qurl, err := emr.sqsQueueManager.QurlFor(emr.emrJobQueue, false)
	if err != nil {
		return nil, errors.Wrap(err, "problem listing queues to poll")
	}
	queues := []string{qurl}
	var runs []RunReceipt
	for _, qurl := range queues {
		//
		// Get new queued Run
		//
		runReceipt, err := emr.sqsQueueManager.ReceiveRun(ctx, qurl)

		if err != nil {
			return runs, errors.Wrapf(err, "problem receiving run from queue url [%s]", qurl)
		}

		if runReceipt.Run == nil {
			continue
		}

		runs = append(runs, RunReceipt{
			RunReceipt:       runReceipt,
			TraceID:          runReceipt.TraceID,
			ParentID:         runReceipt.ParentID,
			SamplingPriority: runReceipt.SamplingPriority,
		})
	}
	return runs, nil
}

func (emr *EMRExecutionEngine) PollStatus(ctx context.Context) (RunReceipt, error) {
	return RunReceipt{}, nil
}

func (emr *EMRExecutionEngine) PollRunStatus(ctx context.Context) (state.Run, error) {
	return state.Run{}, nil
}

func (emr *EMRExecutionEngine) Define(ctx context.Context, td state.Definition) (state.Definition, error) {
	return td, nil
}

func (emr *EMRExecutionEngine) Deregister(ctx context.Context, definition state.Definition) error {
	return errors.Errorf("EMRExecutionEngine does not allow for deregistering of task definitions.")
}

func (emr *EMRExecutionEngine) Get(ctx context.Context, run state.Run) (state.Run, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return run, nil
}

func (emr *EMRExecutionEngine) GetEvents(ctx context.Context, run state.Run) (state.PodEventList, error) {
	var span tracer.Span
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span = utils.TraceJob(ctx, "flotilla.job.emr_get_events", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	return state.PodEventList{}, nil
}

func (emr *EMRExecutionEngine) FetchPodMetrics(ctx context.Context, run state.Run) (state.Run, error) {
	var span tracer.Span
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span = utils.TraceJob(ctx, "flotilla.job.emr_fetch_metrics", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	return run, nil
}

func (emr *EMRExecutionEngine) FetchUpdateStatus(ctx context.Context, run state.Run) (state.Run, error) {
	var span tracer.Span
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, span = utils.TraceJob(ctx, "flotilla.job.emr_fetch_status", run.RunID)
	defer span.Finish()
	utils.TagJobRun(span, run)
	return run, nil
}
func (emr *EMRExecutionEngine) envOverrides(executable state.Executable, run state.Run) []v1.EnvVar {
	pairs := make(map[string]string)
	resources := executable.GetExecutableResources()

	if resources.Env != nil && len(*resources.Env) > 0 {
		for _, ev := range *resources.Env {
			name := emr.sanitizeEnvVar(ev.Name)
			value := ev.Value
			pairs[name] = value
		}
	}

	if run.Env != nil && len(*run.Env) > 0 {
		for _, ev := range *run.Env {
			name := emr.sanitizeEnvVar(ev.Name)
			value := ev.Value
			pairs[name] = value
		}
	}

	var res []v1.EnvVar
	for key := range pairs {
		if len(key) > 0 {
			res = append(res, v1.EnvVar{
				Name:  key,
				Value: pairs[key],
			})
		}
	}

	res = append(res, v1.EnvVar{
		Name: "SPARK_APPLICATION_ID",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				APIVersion: "v1",
				FieldPath:  "metadata.labels['spark-app-selector']",
			},
		},
	})
	return res
}

func (emr *EMRExecutionEngine) sanitizeEnvVar(key string) string {
	// Environment variable can't start with emr $
	if strings.HasPrefix(key, "$") {
		key = strings.Replace(key, "$", "", 1)
	}
	// Environment variable names can't contain spaces.
	key = strings.Replace(key, " ", "", -1)
	return key
}

func (emr *EMRExecutionEngine) constructCmdSlice(command *string) []string {
	cmdString := ""
	if command != nil {
		cmdString = *command
	}
	bashCmd := "bash"
	optLogin := "-l"
	optStr := "-ce"
	return []string{bashCmd, optLogin, optStr, cmdString}
}
