package engine

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEnsureTeamRegistryConfigMap(t *testing.T) {
	ctx := context.Background()

	t.Run("empty team is a no-op", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", "")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("creates configmap when missing", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", "data-engineering")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cm, err := client.CoreV1().ConfigMaps("flotilla-prod").Get(ctx, "flotilla-team-data-engineering", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("configmap not found: %v", err)
		}
		if cm.Data["team"] != "data-engineering" {
			t.Errorf("expected team=data-engineering, got %s", cm.Data["team"])
		}
		if cm.Labels["flotilla.stitchfix.com/team-registry"] != "true" {
			t.Error("missing team-registry label")
		}
	})

	t.Run("updates existing configmap", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		_ = ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", "algorithms")
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", "algorithms")
		if err != nil {
			t.Fatalf("unexpected error on update: %v", err)
		}
	})
}
