package adapter

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EKSAdapter interface {
	AdaptJobToFlotillaRun(job *batchv1.Job, run state.Run) (state.Run, error)
	AdaptFlotillaDefinitionAndRunToJob(td state.Definition, run state.Run) (batchv1.Job, error)
}
type eksAdapter struct{}

//
// NewEKSAdapter configures and returns an eks adapter for translating
// from EKS api specific objects to our representation
//
func NewEKSAdapter(conf config.Config) (EKSAdapter, error) {
	adapter := eksAdapter{}
	return &adapter, nil
}

// TODO: figure this out later.
func (a *eksAdapter) AdaptJobToFlotillaRun(job *batchv1.Job, run state.Run) (state.Run, error) {
	updated := run
	if job.Status.Active == 1 {
		updated.Status = "RUNNING"
	} else if job.Status.Succeeded == 1 {
		var exitCode int64 = 0
		updated.Status = "STOPPED"
		updated.ExitCode = &exitCode
	} else if job.Status.Failed == 1 {
		var exitCode int64 = 1
		updated.Status = "STOPPED"
		updated.ExitCode = &exitCode
	}
	if job != nil && job.Status.StartTime != nil {
		updated.StartedAt = &job.Status.StartTime.Time
	}

	if job != nil && job.Status.CompletionTime != nil {
		updated.FinishedAt = &job.Status.CompletionTime.Time
	}

	return updated, nil
}

func (a *eksAdapter) AdaptFlotillaDefinitionAndRunToJob(definition state.Definition, run state.Run) (batchv1.Job, error) {
	resourceRequirements := a.constructResourceRequirements(definition, run)

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

	nodeLifecycle := state.SpotLifecycle

	if run.NodeLifecycle != nil {
		nodeLifecycle = *run.NodeLifecycle
	}

	lifecycle := "kubernetes.io/lifecycle"
	ttlSecondsAfterFinished := int32(1800)
	activeDeadlineSeconds := int64(86400)
	jobSpec := batchv1.JobSpec{
		TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		ActiveDeadlineSeconds:   &activeDeadlineSeconds,

		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers:    []corev1.Container{container},
				RestartPolicy: corev1.RestartPolicyNever,
				NodeSelector: map[string]string{
					lifecycle: nodeLifecycle,
				},
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

func (a *eksAdapter) constructResourceRequirements(definition state.Definition, run state.Run) corev1.ResourceRequirements {
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

	cpuQuantity := resource.MustParse(fmt.Sprintf("%dm", cpu))
	memoryQuantity := resource.MustParse(fmt.Sprintf("%dMi", mem))

	if definition.Gpu != nil {
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
	return resourceRequirements
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
		name := ev.Name
		value := ev.Value
		pairs[name] = value
	}

	for _, ev := range *run.Env {
		name := ev.Name
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
