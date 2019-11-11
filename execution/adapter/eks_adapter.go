package adapter

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	updated.StartedAt = &job.Status.StartTime.Time
	updated.FinishedAt = &job.Status.CompletionTime.Time
	return updated, nil
}

// TODO: figure what other params are needed.
func (a *eksAdapter) AdaptFlotillaDefinitionAndRunToJob(definition state.Definition, run state.Run) (batchv1.Job, error) {
	limits := make(corev1.ResourceList)
	cpuQuantity := resource.MustParse(fmt.Sprintf("%dm", definition.Cpu))
	if run.Cpu != nil {
		cpuQuantity = resource.MustParse(fmt.Sprintf("%dm", run.Cpu))
	}

	memoryQuantity := resource.MustParse(fmt.Sprintf("%dm", definition.Memory))
	if run.Memory != nil {
		memoryQuantity = resource.MustParse(fmt.Sprintf("%dm", run.Memory))
	}

	if definition.Gpu != nil {
		limits["nvidia.com/gpu"] = resource.MustParse(fmt.Sprintf("%d", definition.Gpu))
	}

	if run.EphemeralStorage != nil {
		limits[corev1.ResourceEphemeralStorage] =
			resource.MustParse(fmt.Sprintf("%dGi", run.EphemeralStorage))
	}

	limits[corev1.ResourceCPU] = cpuQuantity
	limits[corev1.ResourceMemory] = memoryQuantity
	resourceRequirements := corev1.ResourceRequirements{
		Limits: limits,
	}

	container := corev1.Container{
		Name:      run.DefinitionID,
		Image:     run.Image,
		Command:   a.constructCmdSlice(definition.Command),
		Resources: resourceRequirements,
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
					lifecycle: *run.NodeLifecycle,
				},
			},
		},
	}

	eksJob := batchv1.Job{
		Spec: jobSpec,
	}

	return eksJob, nil
}

func (a *eksAdapter) constructCmdSlice(cmdString string) []string {
	bashCmd := "bash"
	optLogin := "-l"
	optStr := "-cex"
	return []string{bashCmd, optLogin, optStr, cmdString}
}
