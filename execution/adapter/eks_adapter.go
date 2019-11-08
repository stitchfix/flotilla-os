package adapter

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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
	// Container spec.
	container := corev1.Container{
		Name:    run.DefinitionID,
		Image:   run.Image,
		Command: a.constructCmdSlice(definition.Command),
	}

	// Job spec.
	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers:    []corev1.Container{container},
				RestartPolicy: "Never",
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
