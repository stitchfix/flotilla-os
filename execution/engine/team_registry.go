package engine

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// attributionKeys are the DIS tag keys written into team-registry ConfigMaps
// so that Kyverno can propagate them to EC2NodeClass instance tags.
var attributionKeys = []string{"team", "cost_center", "cost_center_id", "env", "product_line"}

func ensureTeamRegistryConfigMap(ctx context.Context, client kubernetes.Interface, namespace string, labels map[string]string) error {
	team := labels["team"]
	if team == "" {
		return nil
	}

	data := make(map[string]string, len(attributionKeys))
	for _, k := range attributionKeys {
		if v := labels[k]; v != "" {
			data[k] = v
		}
	}

	name := "flotilla-team-" + team
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":         "flotilla",
				"flotilla.stitchfix.com/team-registry": "true",
			},
		},
		Data: data,
	}

	_, err := client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if k8serrors.IsNotFound(err) {
		_, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	}
	if k8serrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
