package engine

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ensureTeamRegistryConfigMap(ctx context.Context, client kubernetes.Interface, namespace string, team string) error {
	if team == "" {
		return nil
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
		Data: map[string]string{
			"team": team,
		},
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
