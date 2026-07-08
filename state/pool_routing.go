package state

import corev1 "k8s.io/api/core/v1"

type PoolRouting struct {
	RequiredAffinity  *corev1.NodeSelectorRequirement
	PreferredAffinity *corev1.PreferredSchedulingTerm
	Tolerations       []corev1.Toleration
}

func SharedPoolRouting(size string) PoolRouting {
	pool := "standard"
	switch size {
	case "l":
		pool = "large"
	case "xl":
		pool = "xl"
	}

	return PoolRouting{
		RequiredAffinity: &corev1.NodeSelectorRequirement{
			Key:      "flotilla-pool",
			Operator: corev1.NodeSelectorOpIn,
			Values:   []string{pool},
		},
		Tolerations: []corev1.Toleration{
			{Key: "flotilla-pool", Operator: "Equal", Value: pool, Effect: "NoSchedule"},
		},
	}
}
