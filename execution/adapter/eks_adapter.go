package adapter

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
	"time"
)

type EKSAdapter interface {
	AdaptJobToFlotillaRun(job *batchv1.Job, run state.Run, pod *corev1.Pod) (state.Run, error)
	AdaptFlotillaDefinitionAndRunToJob(ctx context.Context, executable state.Executable, run state.Run, schedulerName string, manager state.Manager, araEnabled bool) (batchv1.Job, error)
}
type eksAdapter struct{}

// NewEKSAdapter configures and returns an eks adapter for translating
// from EKS api specific objects to our representation
func NewEKSAdapter() (EKSAdapter, error) {
	adapter := eksAdapter{}
	return &adapter, nil
}

// Adapting Kubernetes batch/v1 job to a Flotilla run object.
// This method maps the exit code & timestamps from Kubernetes to Flotilla's Run object.
func (a *eksAdapter) AdaptJobToFlotillaRun(job *batchv1.Job, run state.Run, pod *corev1.Pod) (state.Run, error) {
	updated := run
	if job.Status.Active == 1 && job.Status.CompletionTime == nil {
		updated.Status = state.StatusRunning
	} else if job.Status.Succeeded == 1 {
		if pod != nil {
			if pod.Status.Phase == corev1.PodSucceeded {
				var exitCode int64 = 0
				var exitReason = fmt.Sprintf("Pod %s Exited Successfully", pod.Name)
				updated.ExitReason = &exitReason
				updated.Status = state.StatusStopped
				updated.ExitCode = &exitCode
			}
		} else {
			var exitCode int64 = 0
			updated.Status = state.StatusStopped
			updated.ExitCode = &exitCode
		}
	} else if job.Status.Failed == 1 {
		var exitCode int64 = 1
		updated.Status = state.StatusStopped
		if pod != nil {
			if pod.Status.ContainerStatuses != nil && len(pod.Status.ContainerStatuses) > 0 {
				containerStatus := pod.Status.ContainerStatuses[len(pod.Status.ContainerStatuses)-1]
				if containerStatus.State.Terminated != nil {
					updated.ExitReason = &containerStatus.State.Terminated.Reason
					exitCode = int64(containerStatus.State.Terminated.ExitCode)
				}
			}
		}
		updated.ExitCode = &exitCode
	}

	if pod != nil && len(pod.Spec.Containers) > 0 {
		container := pod.Spec.Containers[0]
		//First three lines are injected by Flotilla, strip those out.
		if len(container.Command) > 3 {
			cmd := strings.Join(container.Command[3:], "\n")
			updated.Command = &cmd
		}
	}

	if job != nil && job.Status.StartTime != nil {
		updated.StartedAt = &job.Status.StartTime.Time
	}

	if updated.Status == state.StatusStopped {
		if job != nil && job.Status.CompletionTime != nil {
			updated.FinishedAt = &job.Status.CompletionTime.Time
		} else {
			finishedAt := time.Now()
			updated.FinishedAt = &finishedAt
		}
	}
	return updated, nil
}

// Adapting Flotilla run object to Kubernetes batch/v1 job.
// 1. Construction of the cmd that will be run.
// 2. Resources associated to a pod (includes Adaptive Resource Allocation)
// 3. Environment variables to be setup.
// 4. Port mappings.
// 5. Node lifecycle.
// 6. Node affinity and anti-affinity
func (a *eksAdapter) AdaptFlotillaDefinitionAndRunToJob(ctx context.Context, executable state.Executable, run state.Run, schedulerName string, manager state.Manager, araEnabled bool) (batchv1.Job, error) {
	cmd := ""

	if run.Command != nil && len(*run.Command) > 0 {
		cmd = *run.Command
	}

	cmdSlice := a.constructCmdSlice(cmd)
	cmd = strings.Join(cmdSlice[3:], "\n")
	run.Command = &cmd
	resourceRequirements, run := a.constructResourceRequirements(ctx, executable, run, manager, araEnabled)

	volumeMounts, volumes := a.constructVolumeMounts(ctx, executable, run, manager, araEnabled)

	container := corev1.Container{
		Name:            run.RunID,
		Image:           run.Image,
		Command:         cmdSlice,
		Resources:       resourceRequirements,
		Env:             a.envOverrides(executable, run),
		Ports:           a.constructContainerPorts(executable),
		ImagePullPolicy: corev1.PullAlways,
	}

	if volumeMounts != nil {
		container.VolumeMounts = volumeMounts
	}
	affinity := a.constructAffinity(ctx, executable, run, manager)
	tolerations := a.constructTolerations(executable, run)

	annotations := map[string]string{}
	annotations["prometheus.io/port"] = "9090"
	annotations["prometheus.io/scrape"] = "true"

	labels := state.GetLabels(run)

	jobSpec := batchv1.JobSpec{
		TTLSecondsAfterFinished: &state.TTLSecondsAfterFinished,
		ActiveDeadlineSeconds:   run.ActiveDeadlineSeconds,
		BackoffLimit:            &state.EKSBackoffLimit,

		Template: corev1.PodTemplateSpec{
			ObjectMeta: v1.ObjectMeta{
				Annotations: annotations,
				Labels:      labels,
			},
			Spec: corev1.PodSpec{
				SchedulerName:      schedulerName,
				Containers:         []corev1.Container{container},
				RestartPolicy:      corev1.RestartPolicyNever,
				ServiceAccountName: *run.ServiceAccount,
				Affinity:           affinity,
				Tolerations:        tolerations,
			},
		},
	}

	if volumes != nil {
		jobSpec.Template.Spec.Volumes = volumes
	}

	eksJob := batchv1.Job{
		Spec: jobSpec,
		ObjectMeta: v1.ObjectMeta{
			Name: run.RunID,
		},
	}

	return eksJob, nil
}
func (a *eksAdapter) constructEviction(ctx context.Context, run state.Run, manager state.Manager) string {
	if run.Gpu != nil && *run.Gpu > 0 {
		return "false"
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

func (a *eksAdapter) constructContainerPorts(executable state.Executable) []corev1.ContainerPort {
	var containerPorts []corev1.ContainerPort
	executableResources := executable.GetExecutableResources()
	if executableResources.Ports != nil && len(*executableResources.Ports) > 0 {
		for _, port := range *executableResources.Ports {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				ContainerPort: int32(port),
			})
		}
	}
	return containerPorts
}

func (a *eksAdapter) constructTolerations(executable state.Executable, run state.Run) []corev1.Toleration {
	executableResources := executable.GetExecutableResources()
	tolerations := []corev1.Toleration{}

	if (executableResources.Gpu != nil && *executableResources.Gpu > 0) || (run.Gpu != nil && *run.Gpu > 0) {
		toleration := corev1.Toleration{
			Key:      "nvidia.com/gpu",
			Operator: "Equal",
			Value:    "true",
			Effect:   "NoSchedule",
		}
		tolerations = append(tolerations, toleration)
	}
	return tolerations
}

func (a *eksAdapter) constructAffinity(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) *corev1.Affinity {
	affinity := &corev1.Affinity{}
	var requiredMatch []corev1.NodeSelectorRequirement
	var preferredMatches []corev1.PreferredSchedulingTerm
	//todo move to config
	nodeLifecycleKey := "karpenter.sh/capacity-type"
	nodeArchKey := "kubernetes.io/arch"

	//todo remove post migration
	switch run.ClusterName {
	case "flotilla-eks-infra-c":
		nodeLifecycleKey = "node.kubernetes.io/lifecycle"
		nodeArchKey = "kubernetes.io/arch"
	}

	var nodeLifecycle []string
	if run.NodeLifecycle != nil && *run.NodeLifecycle == state.OndemandLifecycle {
		nodeLifecycle = append(nodeLifecycle, "on-demand", "normal")
	} else {
		nodeLifecycle = append(nodeLifecycle, "spot", "on-demand", "normal")
	}

	//todo move to config
	arch := []string{"amd64"}
	if run.Arch != nil && *run.Arch == "arm64" {
		arch = []string{"arm64"}
	}

	requiredMatch = append(requiredMatch, corev1.NodeSelectorRequirement{
		Key:      nodeLifecycleKey,
		Operator: corev1.NodeSelectorOpIn,
		Values:   nodeLifecycle,
	})

	requiredMatch = append(requiredMatch, corev1.NodeSelectorRequirement{
		Key:      nodeArchKey,
		Operator: corev1.NodeSelectorOpIn,
		Values:   arch,
	})

	affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: requiredMatch,
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: preferredMatches,
		},
	}

	return affinity
}

func (a *eksAdapter) constructResourceRequirements(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager, araEnabled bool) (corev1.ResourceRequirements, state.Run) {
	var ephemeralStorageRequestQuantity resource.Quantity
	maxEphemeralStorage := state.MaxEphemeralStorage
	limits := make(corev1.ResourceList)
	requests := make(corev1.ResourceList)

	cpuLimit, memLimit, cpuRequest, memRequest := a.adaptiveResources(ctx, executable, run, manager, araEnabled)

	cpuLimitQuantity := resource.MustParse(fmt.Sprintf("%dm", cpuLimit))
	cpuRequestQuantity := resource.MustParse(fmt.Sprintf("%dm", cpuRequest))

	memLimitQuantity := resource.MustParse(fmt.Sprintf("%dM", memLimit))
	memRequestQuantity := resource.MustParse(fmt.Sprintf("%dM", memRequest))

	limits[corev1.ResourceCPU] = cpuLimitQuantity
	limits[corev1.ResourceMemory] = memLimitQuantity

	requests[corev1.ResourceCPU] = cpuRequestQuantity
	requests[corev1.ResourceMemory] = memRequestQuantity

	executableResources := executable.GetExecutableResources()
	if run.Gpu != nil && *run.Gpu > 0 {
		limits["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", *run.Gpu))
		requests["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", *run.Gpu))
		run.NodeLifecycle = &state.OndemandLifecycle
	} else if executableResources.Gpu != nil && *executableResources.Gpu > 0 {
		limits["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", *executableResources.Gpu))
		requests["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", *executableResources.Gpu))
		run.NodeLifecycle = &state.OndemandLifecycle
	}

	run.Memory = aws.Int64(memRequestQuantity.ScaledValue(resource.Mega))
	run.Cpu = aws.Int64(cpuRequestQuantity.ScaledValue(resource.Milli))
	run.MemoryLimit = aws.Int64(memLimitQuantity.ScaledValue(resource.Mega))
	run.CpuLimit = aws.Int64(cpuLimitQuantity.ScaledValue(resource.Milli))

	if run.EphemeralStorage != nil {
		ephemeralStorageRequest := *run.EphemeralStorage
		if ephemeralStorageRequest > maxEphemeralStorage {
			ephemeralStorageRequest = maxEphemeralStorage
		}
		ephemeralStorageRequestQuantity = resource.MustParse(fmt.Sprintf("%dM", ephemeralStorageRequest))
		requests[corev1.ResourceEphemeralStorage] = ephemeralStorageRequestQuantity
		run.EphemeralStorage = aws.Int64(ephemeralStorageRequestQuantity.ScaledValue(resource.Mega))
	}

	resourceRequirements := corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
	return resourceRequirements, run
}

func (a *eksAdapter) constructVolumeMounts(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager, araEnabled bool) ([]corev1.VolumeMount, []corev1.Volume) {
	var mounts []corev1.VolumeMount = nil
	var volumes []corev1.Volume = nil
	if run.Gpu != nil && *run.Gpu > 0 {
		mounts = make([]corev1.VolumeMount, 1)
		mounts[0] = corev1.VolumeMount{Name: "shared-memory", MountPath: "/dev/shm"}
		volumes = make([]corev1.Volume, 1)
		sharedLimit := resource.MustParse(fmt.Sprintf("%dGi", *run.Gpu*int64(8)))
		emptyDir := corev1.EmptyDirVolumeSource{Medium: "Memory", SizeLimit: &sharedLimit}
		volumes[0] = corev1.Volume{Name: "shared-memory", VolumeSource: corev1.VolumeSource{EmptyDir: &emptyDir}}
	}
	if run.RequiresDocker {
		volumes = append(volumes, corev1.Volume{
			Name: "dockersock",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/var/run/docker.sock",
					Type: nil,
				},
			},
		})
		mounts = append(mounts, corev1.VolumeMount{
			Name:      "dockersock",
			MountPath: "/var/run/docker.sock",
		})
	}
	return mounts, volumes
}

func (a *eksAdapter) adaptiveResources(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager, araEnabled bool) (int64, int64, int64, int64) {
	isGPUJob := run.Gpu != nil && *run.Gpu > 0

	cpuLimit, memLimit := a.getResourceDefaults(run, executable)
	cpuRequest, memRequest := a.getResourceDefaults(run, executable)

	if !isGPUJob {
		estimatedResources, err := manager.EstimateRunResources(ctx, *executable.GetExecutableID(), run.RunID)
		if err == nil {
			cpuRequest = estimatedResources.Cpu
			memRequest = estimatedResources.Memory
		}

		if cpuRequest > cpuLimit {
			cpuLimit = cpuRequest
		}

		if memRequest > memLimit {
			memLimit = memRequest
		}
	}

	cpuRequest, memRequest = a.checkResourceBounds(cpuRequest, memRequest, isGPUJob)
	cpuLimit, memLimit = a.checkResourceBounds(cpuLimit, memLimit, isGPUJob)

	return cpuLimit, memLimit, cpuRequest, memRequest
}

func (a *eksAdapter) checkResourceBounds(cpu int64, mem int64, isGPUJob bool) (int64, int64) {
	maxMem := state.MaxMem
	maxCPU := state.MaxCPU

	if isGPUJob {
		maxMem = state.MaxGPUMem
		maxCPU = state.MaxGPUCPU
	}

	if cpu < state.MinCPU {
		cpu = state.MinCPU
	}
	if cpu > maxCPU {
		cpu = maxCPU
	}

	if mem < state.MinMem {
		mem = state.MinMem
	}
	if mem > maxMem {
		mem = maxMem
	}

	return cpu, mem
}

func (a *eksAdapter) getResourceDefaults(run state.Run, executable state.Executable) (int64, int64) {
	// 1. Init with the global defaults
	cpu := state.MinCPU
	mem := state.MinMem
	executableResources := executable.GetExecutableResources()

	// 2. Look up Run level
	// 3. If not at Run level check Definitions
	if run.Cpu != nil && *run.Cpu != 0 {
		cpu = *run.Cpu
	} else {
		if executableResources.Cpu != nil && *executableResources.Cpu != 0 {
			cpu = *executableResources.Cpu
		}
	}
	if run.Memory != nil && *run.Memory != 0 {
		mem = *run.Memory
	} else {
		if executableResources.Memory != nil && *executableResources.Memory != 0 {
			mem = *executableResources.Memory
		}
	}

	// 4. Override for very large memory requests.
	// Remove after migration.
	if mem >= 36864 && mem < 131072 && (executableResources.Gpu == nil || *executableResources.Gpu == 0) {
		// using the 8x ratios between cpu and memory ~ r5 class of instances
		cpuOverride := mem / 8
		if cpuOverride > cpu {
			cpu = cpuOverride
		}
	}

	return cpu, mem
}

func (a *eksAdapter) getLastRun(ctx context.Context, manager state.Manager, run state.Run) state.Run {
	var lastRun state.Run
	runList, err := manager.ListRuns(ctx, 1, 0, "started_at", "desc", map[string][]string{
		"queued_at_since": {
			time.Now().AddDate(0, 0, -7).Format(time.RFC3339),
		},
		"status":        {state.StatusStopped},
		"command":       {strings.Replace(*run.Command, "'", "''", -1)},
		"executable_id": {*run.ExecutableID},
	}, nil, []string{state.EKSEngine})
	if err == nil && len(runList.Runs) > 0 {
		lastRun = runList.Runs[0]
	}
	return lastRun
}

func (a *eksAdapter) constructCmdSlice(cmdString string) []string {
	bashCmd := "bash"
	optLogin := "-l"
	optStr := "-cex"
	return []string{bashCmd, optLogin, optStr, cmdString}
}

func (a *eksAdapter) envOverrides(executable state.Executable, run state.Run) []corev1.EnvVar {
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

	var res []corev1.EnvVar
	for key := range pairs {
		if len(key) > 0 {
			res = append(res, corev1.EnvVar{
				Name:  key,
				Value: pairs[key],
			})
		}
	}
	return res
}

func (a *eksAdapter) sanitizeEnvVar(key string) string {
	// Environment variable can't start with a $
	if strings.HasPrefix(key, "$") {
		key = strings.Replace(key, "$", "", 1)
	}
	// Environment variable names can't contain spaces.
	key = strings.Replace(key, " ", "", -1)
	return key
}

func (a *eksAdapter) sanitizeLabel(key string) string {
	key = strings.TrimSpace(key)
	key = regexp.MustCompile(`[^-a-z0-9A-Z_.]+`).ReplaceAllString(key, "_")
	key = strings.TrimPrefix(key, "_")
	key = strings.ToLower(key)
	if len(key) > 63 {
		key = key[:63]
	}
	return key
}
