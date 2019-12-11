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
	AdaptFlotillaDefinitionAndRunToJob(td state.Definition, run state.Run, sa string, schedulerName string, manager state.Manager) (batchv1.Job, error)
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


func (a *eksAdapter) AdaptFlotillaDefinitionAndRunToJob(definition state.Definition, run state.Run, sa string, schedulerName string, manager state.Manager) (batchv1.Job, error) {
	cmd := definition.Command
	if run.Command != nil {
		cmd = *run.Command
	}
	run.Command = &cmd
	resourceRequirements := a.constructResourceRequirements(definition, run, manager)

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

func (a *eksAdapter) constructResourceRequirements(definition state.Definition, run state.Run, manager state.Manager) corev1.ResourceRequirements {
	limits := make(corev1.ResourceList)
	cpu, mem := a.adaptiveResources(definition, run, manager)
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

func (a *eksAdapter) adaptiveResources(definition state.Definition, run state.Run, manager state.Manager) (int64, int64) {
	cpu := state.MinCPU
	mem := state.MinMem

	if definition.AdaptiveResourceAllocation != nil && *definition.AdaptiveResourceAllocation == true {
		pastRuns, err := manager.ListRuns(1, 0, "started_at", "desc", map[string][]string{
			"queued_at_since": {
				time.Now().AddDate(0, 0, -30).Format(time.RFC3339),
			},
			"status":        {state.StatusStopped},
			"command":       {*run.Command},
			"definition_id": {definition.DefinitionID},
		}, nil, []string{state.EKSEngine})

		if err != nil {
			return cpu, mem
		}

		if len(pastRuns.Runs) > 0 {
			lastRun := pastRuns.Runs[0]
			if lastRun.MaxMemoryUsed != nil && lastRun.MaxCpuUsed != nil {
				if *lastRun.ExitCode == 0 {
					cpu = int64(float64(*lastRun.MaxCpuUsed) * 1.1)
					mem = int64(float64(*lastRun.MaxMemoryUsed) * 1.25)
				} else {
					if !strings.Contains(*lastRun.ExitReason, "OOM") {
						cpu = int64(float64(*lastRun.Cpu) * 1.1)
						mem = int64(float64(*lastRun.Memory) * 1.50)
					} else {
						cpu = *lastRun.Cpu
						mem = *lastRun.Memory
					}
				}
			}
		}
	}

	if cpu == state.MinCPU {
		if run.Cpu != nil && *run.Cpu != 0 {
			cpu = *run.Cpu
		} else {
			if definition.Cpu != nil && *definition.Cpu != 0 {
				cpu = *definition.Cpu
			}
		}
	}

	if mem == state.MinMem {
		if run.Memory != nil && *run.Memory != 0 {
			mem = *run.Memory
		} else {
			if definition.Memory != nil && *definition.Memory != 0 {
				mem = *definition.Memory
			}
		}
	}

	return cpu, mem
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
