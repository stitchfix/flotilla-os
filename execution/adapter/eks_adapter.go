package adapter

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type EKSAdapter interface {
	AdaptJobToFlotillaRun(job *batchv1.Job, run state.Run, pod *corev1.Pod) (state.Run, error)
	AdaptFlotillaDefinitionAndRunToJob(td state.Definition, run state.Run, sa string, schedulerName string) (batchv1.Job, error)
}
type eksAdapter struct{}

//
// NewEKSAdapter configures and returns an eks adapter for translating
// from EKS api specific objects to our representation
//
func NewEKSAdapter() (EKSAdapter, error) {
	adapter := eksAdapter{}

	return &adapter, nil
}

func (a *eksAdapter) AdaptJobToFlotillaRun(job *batchv1.Job, run state.Run, pod *corev1.Pod) (state.Run, error) {
	updated := run
	if job.Status.Active == 1 && job.Status.CompletionTime == nil {
		updated.Status = state.StatusRunning
	} else if job.Status.Succeeded == 1 {
		var exitCode int64 = 0
		updated.Status = state.StatusStopped
		updated.ExitCode = &exitCode
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

func (a *eksAdapter) AdaptFlotillaDefinitionAndRunToJob(definition state.Definition, run state.Run, sa string, schedulerName string) (batchv1.Job, error) {
	resourceRequirements, run := a.constructResourceRequirements(definition, run)

	cmd := definition.Command
	if run.Command != nil {
		cmd = *run.Command
	}

	container := corev1.Container{
		Name:      run.RunID,
		Image:     run.Image,
		Command:   a.constructCmdSlice(cmd),
		Resources: resourceRequirements,
		Env:       a.envOverrides(definition, run),
	}

	affinity := a.constructAffinity(definition, run)
	annotations := map[string]string{"cluster-autoscaler.kubernetes.io/safe-to-evict": "false"}

	activeDeadlineSeconds := state.SpotActiveDeadlineSeconds
	if *run.NodeLifecycle == state.OndemandLifecycle {
		activeDeadlineSeconds = state.OndemandActiveDeadlineSeconds
	}

	jobSpec := batchv1.JobSpec{
		TTLSecondsAfterFinished: &state.TTLSecondsAfterFinished,
		ActiveDeadlineSeconds:   &activeDeadlineSeconds,
		BackoffLimit:            &state.EKSBackoffLimit,

		Template: corev1.PodTemplateSpec{
			ObjectMeta: v1.ObjectMeta{
				Annotations: annotations,
			},
			Spec: corev1.PodSpec{
				SchedulerName:      schedulerName,
				Containers:         []corev1.Container{container},
				RestartPolicy:      corev1.RestartPolicyNever,
				ServiceAccountName: sa,
				Affinity:           affinity,
			},
		},
	}

	eksJob := batchv1.Job{
		Spec: jobSpec,
		ObjectMeta: v1.ObjectMeta{
			Name: run.RunID,
		},
	}

	return eksJob, nil
}

func (a *eksAdapter) constructAffinity(definition state.Definition, run state.Run) *corev1.Affinity {
	affinity := &corev1.Affinity{}
	var requiredMatch []corev1.NodeSelectorRequirement

	gpuNodeTypes := []string{"p3.2xlarge", "p3.8xlarge", "p3.16xlarge"}
	cpuNodeTypes := []string{"c5.2xlarge", "c5.4xlarge", "c5.9xlarge"}

	var nodeLifecycle []string
	if *run.NodeLifecycle == state.OndemandLifecycle {
		nodeLifecycle = append(nodeLifecycle, "normal")
	} else {
		nodeLifecycle = append(nodeLifecycle, "spot")
	}

	if definition.Gpu == nil || *definition.Gpu <= 0 {
		requiredMatch = append(requiredMatch, corev1.NodeSelectorRequirement{
			Key:      "beta.kubernetes.io/instance-type",
			Operator: corev1.NodeSelectorOpNotIn,
			Values:   gpuNodeTypes,
		})

		//For high cpu jobs - assign to c5 node types.
		if run.Memory != nil &&
			run.Cpu != nil &&
			*run.Cpu > int64(0) &&
			*run.Memory > int64(0) &&
			float64(*run.Cpu)/float64(*run.Memory) >= 0.5 {
			requiredMatch = append(requiredMatch, corev1.NodeSelectorRequirement{
				Key:      "beta.kubernetes.io/instance-type",
				Operator: corev1.NodeSelectorOpIn,
				Values:   cpuNodeTypes,
			})
		}

	}

	requiredMatch = append(requiredMatch, corev1.NodeSelectorRequirement{
		Key:      "kubernetes.io/lifecycle",
		Operator: corev1.NodeSelectorOpIn,
		Values:   nodeLifecycle,
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
		},
	}

	return affinity
}

func (a *eksAdapter) constructResourceRequirements(definition state.Definition, run state.Run) (corev1.ResourceRequirements, state.Run) {
	limits := make(corev1.ResourceList)
	cpu := *definition.Cpu
	if run.Cpu != nil {
		cpu = *run.Cpu
	}
	if cpu < state.MinCPU {
		cpu = state.MinCPU
	}

	mem := *definition.Memory
	if run.Memory != nil {
		mem = *run.Memory
	}
	if mem < state.MinMem {
		mem = state.MinMem
	}

	// CPU Override for legacy jobs that are high memory, remove once migration is complete.
	if mem >= 36864 && mem < 131072 && (definition.Gpu == nil || *definition.Gpu == 0) {
		// using the 8x ratios between cpu and memory ~ r5 class of instances
		cpuOverride := mem / 8
		if cpuOverride > cpu {
			cpu = cpuOverride
		}
	}

	cpuQuantity := resource.MustParse(fmt.Sprintf("%dm", cpu))
	assignedCpu := cpuQuantity.ScaledValue(resource.Milli)
	run.Cpu = &assignedCpu

	memoryQuantity := resource.MustParse(fmt.Sprintf("%dM", mem))
	assignedMem := memoryQuantity.ScaledValue(resource.Mega)
	run.Memory = &assignedMem

	if definition.Gpu != nil && *definition.Gpu > 0 {
		limits["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", *definition.Gpu))
	}
	if run.EphemeralStorage != nil {
		limits[corev1.ResourceEphemeralStorage] =
			resource.MustParse(fmt.Sprintf("%dGi", *run.EphemeralStorage))
	}
	limits[corev1.ResourceCPU] = cpuQuantity
	limits[corev1.ResourceMemory] = memoryQuantity
	resourceRequirements := corev1.ResourceRequirements{
		Limits: limits,
	}
	return resourceRequirements, run
}

func (a *eksAdapter) constructCmdSlice(cmdString string) []string {
	bashCmd := "bash"
	optLogin := "-l"
	optStr := "-cex"
	return []string{bashCmd, optLogin, optStr, cmdString}
}

func (a *eksAdapter) envOverrides(definition state.Definition, run state.Run) []corev1.EnvVar {
	pairs := make(map[string]string)
	for _, ev := range *definition.Env {
		name := a.sanitizeEnvVar(ev.Name)
		value := ev.Value
		pairs[name] = value
	}

	for _, ev := range *run.Env {
		name := a.sanitizeEnvVar(ev.Name)
		value := ev.Value
		pairs[name] = value
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
