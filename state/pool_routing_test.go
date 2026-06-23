package state

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestSharedPoolRouting_Small(t *testing.T) {
	r := SharedPoolRouting("s")
	assertRequiredPool(t, r, "standard")
	assertTolerationValues(t, r.Tolerations, []string{"standard"})
}

func TestSharedPoolRouting_Medium(t *testing.T) {
	r := SharedPoolRouting("m")
	assertRequiredPool(t, r, "standard")
	assertTolerationValues(t, r.Tolerations, []string{"standard"})
}

func TestSharedPoolRouting_Large(t *testing.T) {
	r := SharedPoolRouting("l")
	assertRequiredPool(t, r, "large")
	assertTolerationValues(t, r.Tolerations, []string{"large"})
}

func TestSharedPoolRouting_XL(t *testing.T) {
	r := SharedPoolRouting("xl")
	assertRequiredPool(t, r, "xl")
	assertTolerationValues(t, r.Tolerations, []string{"xl"})
}

func TestSharedPoolRouting_Unknown(t *testing.T) {
	r := SharedPoolRouting("unknown")
	assertRequiredPool(t, r, "standard")
}

func TestSharedPoolRouting_SmallAndMediumAreIdentical(t *testing.T) {
	s := SharedPoolRouting("s")
	m := SharedPoolRouting("m")
	if s.RequiredAffinity.Values[0] != m.RequiredAffinity.Values[0] {
		t.Errorf("small and medium should route to same pool")
	}
}

func TestSharedPoolRouting_NoCrossPoolTolerations(t *testing.T) {
	for _, size := range []string{"s", "m", "l", "xl"} {
		r := SharedPoolRouting(size)
		if len(r.Tolerations) != 1 {
			t.Errorf("size %q: expected exactly 1 toleration, got %d", size, len(r.Tolerations))
		}
	}
}

func TestSharedPoolRouting_NoPreferredAffinity(t *testing.T) {
	for _, size := range []string{"s", "m", "l", "xl"} {
		r := SharedPoolRouting(size)
		if r.PreferredAffinity != nil {
			t.Errorf("size %q: should not have preferred affinity", size)
		}
	}
}

func assertRequiredPool(t *testing.T, r PoolRouting, expected string) {
	t.Helper()
	if r.RequiredAffinity == nil {
		t.Fatal("expected required affinity")
	}
	if r.RequiredAffinity.Key != "flotilla-pool" {
		t.Errorf("required key should be flotilla-pool, got %s", r.RequiredAffinity.Key)
	}
	if r.RequiredAffinity.Values[0] != expected {
		t.Errorf("required value should be %s, got %v", expected, r.RequiredAffinity.Values)
	}
}

func assertTolerationValues(t *testing.T, tolerations []corev1.Toleration, expected []string) {
	t.Helper()
	if len(tolerations) != len(expected) {
		t.Errorf("expected %d tolerations, got %d", len(expected), len(tolerations))
		return
	}
	for i, tol := range tolerations {
		if tol.Key != "flotilla-pool" {
			t.Errorf("toleration[%d] key should be flotilla-pool, got %s", i, tol.Key)
		}
		if tol.Value != expected[i] {
			t.Errorf("toleration[%d] value should be %s, got %s", i, expected[i], tol.Value)
		}
		if tol.Effect != "NoSchedule" {
			t.Errorf("toleration[%d] effect should be NoSchedule, got %s", i, tol.Effect)
		}
	}
}
