package adapter

import (
	"strings"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1")

type EKSAdapter interface {
	AdaptJobToFlotillaRun(job *batchv1.Job) (state.Run, error)
	AdaptFlotillaRunToJob(sr *state.StatelessRun) (batchv1.Job, error)
	AdaptCommand(cmd string) []string
}
type eksAdapter struct {}

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
func (a *eksAdapter) AdaptFlotillaRunToJob(sr *state.StatelessRun) (batchv1.Job, error) {
	// Container spec.
	container := corev1.Container{
		Name:    sr.Name,
		Image:   sr.Image,
		Command: a.AdaptCommand(sr.Command),
	}

	// Template.
	template := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers:    []corev1.Container{container},
			RestartPolicy: "Never",
		},
	}

	// Job spec.
	jobspec := batchv1.JobSpec{
		Template:     template,
	}

	runID, err := state.NewRunID()

	if err != nil {
		return batchv1.Job{}, nil
	}

	jobMetadata := metav1.ObjectMeta{
		Name: runID,
	}

	eksJob := batchv1.Job{
		ObjectMeta: jobMetadata,
		Spec:       jobspec,
	}

	return eksJob, nil
}

func (a *eksAdapter) AdaptCommand(cmd string) []string {
	return strings.Split(cmd, "\n")
}