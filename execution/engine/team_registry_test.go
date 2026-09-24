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
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", map[string]string{})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("creates configmap when missing", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		labels := map[string]string{
			"team": "data-engineering",
		}
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", labels)
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
		labels := map[string]string{"team": "algorithms"}
		_ = ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", labels)
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", labels)
		if err != nil {
			t.Fatalf("unexpected error on update: %v", err)
		}
	})

	t.Run("writes full attribution tags into configmap data", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		labels := map[string]string{
			"team":           "data-engineering",
			"cost_center":    "Engineering",
			"cost_center_id": "1234",
			"env":            "production",
			"product_line":   "platform",
		}
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", labels)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cm, err := client.CoreV1().ConfigMaps("flotilla-prod").Get(ctx, "flotilla-team-data-engineering", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("configmap not found: %v", err)
		}

		expected := map[string]string{
			"team":           "data-engineering",
			"cost_center":    "Engineering",
			"cost_center_id": "1234",
			"env":            "production",
			"product_line":   "platform",
		}
		for k, want := range expected {
			if got := cm.Data[k]; got != want {
				t.Errorf("Data[%q] = %q, want %q", k, got, want)
			}
		}
	})

	t.Run("omits empty attribution labels from data", func(t *testing.T) {
		client := fake.NewSimpleClientset()
		labels := map[string]string{
			"team":           "algorithms",
			"cost_center":    "Research",
			"cost_center_id": "",
			"env":            "",
			"product_line":   "algo",
		}
		err := ensureTeamRegistryConfigMap(ctx, client, "flotilla-prod", labels)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		cm, err := client.CoreV1().ConfigMaps("flotilla-prod").Get(ctx, "flotilla-team-algorithms", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("configmap not found: %v", err)
		}

		// Present keys
		if cm.Data["team"] != "algorithms" {
			t.Errorf("Data[team] = %q, want %q", cm.Data["team"], "algorithms")
		}
		if cm.Data["cost_center"] != "Research" {
			t.Errorf("Data[cost_center] = %q, want %q", cm.Data["cost_center"], "Research")
		}
		if cm.Data["product_line"] != "algo" {
			t.Errorf("Data[product_line] = %q, want %q", cm.Data["product_line"], "algo")
		}

		// Empty-value keys must be absent
		for _, k := range []string{"cost_center_id", "env"} {
			if v, ok := cm.Data[k]; ok {
				t.Errorf("Data[%q] should be absent but got %q", k, v)
			}
		}
	})
}
