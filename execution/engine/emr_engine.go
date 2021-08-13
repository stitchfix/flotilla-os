package engine

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/clients/metrics"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
	v1 "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sJson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"strings"
)

//
// EMRExecutionEngine submits runs to EMR-EKS.
//
type EMRExecutionEngine struct {
	sqsQueueManager     queue.Manager
	log                 flotillaLog.Logger
	emrJobStatusQueue   string
	emrJobQueue         string
	emrJobNamespace     string
	emrJobRoleArn       string
	emrJobSA            string
	emrVirtualCluster   string
	emrContainersClient *emrcontainers.EMRContainers
	s3Client            *s3.S3
	awsRegion           string
	s3LogsBucket        string
	s3LogsBasePath      string
	s3ManifestBucket    string
	s3ManifestBasePath  string
	serializer          *k8sJson.Serializer
}

//
// Initialize configures the EMRExecutionEngine and initializes internal clients
//
func (ee *EMRExecutionEngine) Initialize(conf config.Config) error {

	ee.emrVirtualCluster = conf.GetString("emr.virtual_cluster")
	ee.emrJobQueue = conf.GetString("emr.job_queue")
	ee.emrJobStatusQueue = conf.GetString("emr.job_status_queue")
	ee.emrJobNamespace = conf.GetString("emr.job_namespace")
	ee.emrJobRoleArn = conf.GetString("emr.job_role_arn")
	ee.awsRegion = conf.GetString("emr.aws_region")
	ee.s3LogsBucket = conf.GetString("emr.log.bucket")
	ee.s3LogsBasePath = conf.GetString("emr.log.base_path")
	ee.s3ManifestBucket = conf.GetString("emr.manifest.bucket")
	ee.s3ManifestBasePath = conf.GetString("emr.manifest.base_path")
	ee.emrJobSA = conf.GetString("eks.service_account")

	awsConfig := &aws.Config{Region: aws.String(ee.awsRegion)}
	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: *awsConfig}))
	ee.s3Client = s3.New(sess, aws.NewConfig().WithRegion(ee.awsRegion))
	ee.emrContainersClient = emrcontainers.New(sess, aws.NewConfig().WithRegion(ee.awsRegion))

	ee.serializer = k8sJson.NewSerializerWithOptions(
		k8sJson.DefaultMetaFactory, nil, nil,
		k8sJson.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		},
	)
	return nil
}

func (ee *EMRExecutionEngine) Execute(executable state.Executable, run state.Run, manager state.Manager) (state.Run, bool, error) {
	startJobRunInput := emrcontainers.StartJobRunInput{
		ClientToken: &run.RunID,
		ConfigurationOverrides: &emrcontainers.ConfigurationOverrides{
			MonitoringConfiguration: &emrcontainers.MonitoringConfiguration{
				PersistentAppUI: aws.String(emrcontainers.PersistentAppUIEnabled),
				S3MonitoringConfiguration: &emrcontainers.S3MonitoringConfiguration{
					LogUri: aws.String(fmt.Sprintf("%s/%s", ee.s3LogsBucket, ee.s3LogsBasePath)),
				},
			},
			ApplicationConfiguration: []*emrcontainers.Configuration{
				{
					Classification: aws.String("spark-defaults"),
					Properties: map[string]*string{
						"spark.kubernetes.driver.podTemplateFile":   ee.driverPodTemplate(executable, run, manager),
						"spark.kubernetes.executor.podTemplateFile": ee.executorPodTemplate(executable, run, manager),
						"spark.kubernetes.container.image":          &run.Image},
				},
			},
		},
		ExecutionRoleArn: &ee.emrJobRoleArn,
		JobDriver: &emrcontainers.JobDriver{
			SparkSubmitJobDriver: &emrcontainers.SparkSubmitJobDriver{
				EntryPoint:            run.SparkExtension.SparkSubmitJobDriver.EntryPoint,
				EntryPointArguments:   run.SparkExtension.SparkSubmitJobDriver.EntryPointArguments,
				SparkSubmitParameters: ee.sparkSubmitParams(run),
			}},
		Name:             &run.RunID,
		ReleaseLabel:     run.SparkExtension.EMRReleaseLabel,
		VirtualClusterId: &ee.emrVirtualCluster,
	}

	key := aws.String(fmt.Sprintf("%s/%s/%s.yaml", ee.s3ManifestBasePath, run.RunID, "start-job-run-input"))
	ee.writeStringToS3(aws.String(startJobRunInput.String()), key)

	startJobRunOutput, err := ee.emrContainersClient.StartJobRun(&startJobRunInput)
	if err == nil {
		run.SparkExtension.VirtualClusterId = startJobRunOutput.VirtualClusterId
		run.SparkExtension.EMRJobId = startJobRunOutput.Id
		run.Status = state.StatusPending
		_ = metrics.Increment(metrics.EngineEMRExecute, []string{string(metrics.StatusSuccess)}, 1)
	} else {
		run.ExitReason = aws.String("Failed to submit job to EMR/EKS.")
		_ = ee.log.Log("EMR job submission error", "error", err.Error())
		_ = metrics.Increment(metrics.EngineEKSExecute, []string{string(metrics.StatusFailure)}, 1)
		return run, false, err
	}
	return run, false, nil
}

func (ee *EMRExecutionEngine) driverPodTemplate(executable state.Executable, run state.Run, manager state.Manager) *string {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
				"flotilla-run-id": run.RunID},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{{
				Name: "shared-lib-volume",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &(v1.EmptyDirVolumeSource{}),
				},
			}},
			Containers: []v1.Container{
				{
					Name: "spark-kubernetes-driver",
					Env:  ee.envOverrides(executable, run),
				},
			},
			InitContainers: []v1.Container{{
				Name:  fmt.Sprintf("init-driver-%s", run.RunID),
				Image: run.Image,
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "shared-lib-volume",
						MountPath: "/var/lib/app",
					},
				},
			}},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: ee.emrJobSA,
			Affinity:           ee.constructAffinity(executable, run, manager),
		},
	}

	key := aws.String(fmt.Sprintf("%s/%s/%s.yaml", ee.s3ManifestBasePath, run.RunID, "driver-template"))
	ee.writeK8ObjToS3(&pod, key)
	return key
}

func (ee *EMRExecutionEngine) executorPodTemplate(executable state.Executable, run state.Run, manager state.Manager) *string {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
				"flotilla-run-id": run.RunID},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{{
				Name: "shared-lib-volume",
				VolumeSource: v1.VolumeSource{
					EmptyDir: &(v1.EmptyDirVolumeSource{}),
				},
			}},
			Containers: []v1.Container{
				{
					Name:  "spark-kubernetes-executor",
					Image: run.Image,
					Env:   ee.envOverrides(executable, run),
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "shared-lib-volume",
							MountPath: "/var/lib/app",
						},
					},
				},
			},
			InitContainers: []v1.Container{{
				Name:  fmt.Sprintf("init-executor-%s", run.RunID),
				Image: run.Image,
				VolumeMounts: []v1.VolumeMount{
					{
						Name:      "shared-lib-volume",
						MountPath: "/var/lib/app",
					},
				},
			}},
			RestartPolicy:      v1.RestartPolicyNever,
			ServiceAccountName: ee.emrJobSA,
			Affinity:           ee.constructAffinity(executable, run, manager),
		},
	}

	key := aws.String(fmt.Sprintf("%s/%s/%s.yaml", ee.s3ManifestBasePath, run.RunID, "executor-template"))
	ee.writeK8ObjToS3(&pod, key)
	return key
}

func (ee *EMRExecutionEngine) writeK8ObjToS3(obj runtime.Object, key *string) {
	var b0 bytes.Buffer
	err := ee.serializer.Encode(obj, &b0)

	if err == nil {
		putObject := s3.PutObjectInput{
			Bucket:      aws.String(ee.s3ManifestBucket),
			Body:        bytes.NewReader(b0.Bytes()),
			Key:         key,
			ContentType: aws.String("text/yaml"),
		}
		_, err = ee.s3Client.PutObject(&putObject)
		if err != nil {
			_ = ee.log.Log("s3_upload_error", "error", err.Error())
		}
	}
}

func (ee *EMRExecutionEngine) writeStringToS3(key *string, body *string) {
	if body != nil && key != nil {
		putObject := s3.PutObjectInput{
			Bucket:      aws.String(ee.s3ManifestBucket),
			Body:        strings.NewReader(*body),
			Key:         key,
			ContentType: aws.String("text/yaml"),
		}
		_, err := ee.s3Client.PutObject(&putObject)
		if err != nil {
			_ = ee.log.Log("s3_upload_error", "error", err.Error())
		}
	}
}

func (a *EMRExecutionEngine) constructAffinity(executable state.Executable, run state.Run, manager state.Manager) *v1.Affinity {
	affinity := &v1.Affinity{}
	executableResources := executable.GetExecutableResources()
	var requiredMatch []v1.NodeSelectorRequirement

	gpuNodeTypes := []string{"p3.2xlarge", "p3.8xlarge", "p3.16xlarge"}

	var nodeLifecycle []string
	if run.NodeLifecycle != nil && *run.NodeLifecycle == state.OndemandLifecycle {
		nodeLifecycle = append(nodeLifecycle, "normal")
	} else {
		nodeLifecycle = append(nodeLifecycle, "spot")
	}

	if (executableResources.Gpu == nil || *executableResources.Gpu <= 0) && (run.Gpu == nil || *run.Gpu <= 0) {
		requiredMatch = append(requiredMatch, v1.NodeSelectorRequirement{
			Key:      "beta.kubernetes.io/instance-type",
			Operator: v1.NodeSelectorOpNotIn,
			Values:   gpuNodeTypes,
		})

		nodeList, err := manager.ListFailingNodes()

		if err == nil && len(nodeList) > 0 {
			requiredMatch = append(requiredMatch, v1.NodeSelectorRequirement{
				Key:      "kubernetes.io/hostname",
				Operator: v1.NodeSelectorOpNotIn,
				Values:   nodeList,
			})
		}
	}

	requiredMatch = append(requiredMatch, v1.NodeSelectorRequirement{
		Key:      "node.kubernetes.io/lifecycle",
		Operator: v1.NodeSelectorOpIn,
		Values:   nodeLifecycle,
	})

	affinity = &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: requiredMatch,
					},
				},
			},
		},
	}
	return affinity
}

func (ee *EMRExecutionEngine) sparkSubmitParams(run state.Run) *string {
	var buffer bytes.Buffer
	for _, k := range run.SparkExtension.SparkSubmitJobDriver.SparkSubmitConf {
		buffer.WriteString(fmt.Sprintf(" --conf %s=%s", *k.Name, *k.Value))
	}
	return aws.String(buffer.String())
}

func (ee *EMRExecutionEngine) Terminate(run state.Run) error {
	cancelJobRunInput := emrcontainers.CancelJobRunInput{
		Id:               run.SparkExtension.EMRJobId,
		VirtualClusterId: run.SparkExtension.VirtualClusterId,
	}
	_, err := ee.emrContainersClient.CancelJobRun(&cancelJobRunInput)
	if err != nil {
		_ = metrics.Increment(metrics.EngineEMRTerminate, []string{string(metrics.StatusFailure)}, 1)
		_ = ee.log.Log("EMR job termination error", "error", err.Error())
	}
	_ = metrics.Increment(metrics.EngineEMRTerminate, []string{string(metrics.StatusSuccess)}, 1)
	return err
}

func (ee *EMRExecutionEngine) Enqueue(run state.Run) error {
	qurl, err := ee.sqsQueueManager.QurlFor(ee.emrJobQueue, false)
	if err != nil {
		_ = metrics.Increment(metrics.EngineEMREnqueue, []string{string(metrics.StatusFailure)}, 1)
		_ = ee.log.Log("EMR job enqueue error", "error", err.Error())
		return errors.Wrapf(err, "problem getting queue url for [%s]", run.ClusterName)
	}

	// Queue run
	if err = ee.sqsQueueManager.Enqueue(qurl, run); err != nil {
		_ = metrics.Increment(metrics.EngineEMREnqueue, []string{string(metrics.StatusFailure)}, 1)
		_ = ee.log.Log("EMR job enqueue error", "error", err.Error())
		return errors.Wrapf(err, "problem enqueing run [%s] to queue [%s]", run.RunID, qurl)
	}

	_ = metrics.Increment(metrics.EngineEMREnqueue, []string{string(metrics.StatusSuccess)}, 1)
	return nil
}

func (ee *EMRExecutionEngine) PollRuns() ([]RunReceipt, error) {
	qurl, err := ee.sqsQueueManager.QurlFor(ee.emrJobQueue, false)
	if err != nil {
		return nil, errors.Wrap(err, "problem listing queues to poll")
	}
	queues := []string{qurl}
	var runs []RunReceipt
	for _, qurl := range queues {
		//
		// Get new queued Run
		//
		runReceipt, err := ee.sqsQueueManager.ReceiveRun(qurl)

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

func (ee *EMRExecutionEngine) PollStatus() (RunReceipt, error) {
	return RunReceipt{}, nil
}

func (ee *EMRExecutionEngine) PollRunStatus() (state.Run, error) {
	return state.Run{}, nil
}

func (ee *EMRExecutionEngine) Define(td state.Definition) (state.Definition, error) {
	return td, nil
}

func (ee *EMRExecutionEngine) Deregister(definition state.Definition) error {
	return errors.Errorf("EMRExecutionEngine does not allow for deregistering of task definitions.")
}

func (ee *EMRExecutionEngine) Get(run state.Run) (state.Run, error) {
	return run, nil
}

func (ee *EMRExecutionEngine) GetEvents(run state.Run) (state.PodEventList, error) {
	return state.PodEventList{}, nil
}

func (ee *EMRExecutionEngine) FetchPodMetrics(run state.Run) (state.Run, error) {
	return run, nil
}

func (ee *EMRExecutionEngine) FetchUpdateStatus(run state.Run) (state.Run, error) {
	return run, nil
}
func (a *EMRExecutionEngine) envOverrides(executable state.Executable, run state.Run) []v1.EnvVar {
	pairs := make(map[string]string)
	resources := executable.GetExecutableResources()

	if resources.Env != nil && len(*resources.Env) > 0 {
		for _, ev := range *resources.Env {
			name := a.sanitizeEnvVar(ev.Name)
			value := ev.Value
			pairs[name] = value
		}
	}

	if run.Env != nil && len(*run.Env) > 0 {
		for _, ev := range *run.Env {
			name := a.sanitizeEnvVar(ev.Name)
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
	return res
}

func (a *EMRExecutionEngine) sanitizeEnvVar(key string) string {
	// Environment variable can't start with a $
	if strings.HasPrefix(key, "$") {
		key = strings.Replace(key, "$", "", 1)
	}
	// Environment variable names can't contain spaces.
	key = strings.Replace(key, " ", "", -1)
	return key
}
