package adapter

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type EKSAdapter interface {
	AdaptJobToFlotillaRun(job *batchv1.Job) (state.Run, error)
	AdaptFlotillaDefinitionAndRunToJob(td *state.Definition, run *state.Run) (batchv1.Job, error)
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
func (a *eksAdapter) AdaptJobToFlotillaRun(job *batchv1.Job) (state.Run, error) {
	run := state.Run{}
	return run, nil
}

// TODO: figure what other params are needed.
func (a *eksAdapter) AdaptFlotillaDefinitionAndRunToJob(td *state.Definition, run *state.Run) (batchv1.Job, error) {
	// Container spec.
	container := corev1.Container{
		Name:    td.DefinitionID,
		Image:   run.Image,
		Command: strings.Split(*run.Command, "\n"),
	}

	// Job spec.
	jobspec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers:    []corev1.Container{container},
				RestartPolicy: "Never",
			},
		},
	}

	eksJob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: run.RunID,
		},
		Spec: jobspec,
	}

	return eksJob, nil
}
