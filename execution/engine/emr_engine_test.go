package engine

import (
	"fmt"
	"testing"

	"github.com/stitchfix/flotilla-os/state"
)

type nopLogger struct{}

func (nopLogger) Log(_ ...interface{}) error  { return nil }
func (nopLogger) Event(_ ...interface{}) error { return nil }

func newTestEMREngine() *EMRExecutionEngine {
	return &EMRExecutionEngine{log: nopLogger{}}
}

func TestEmrJobRunTags_ValidLabels(t *testing.T) {
	emr := newTestEMREngine()
	labels := state.Labels{
		"team":           "data-platform",
		"cost_center":    "CC-1234",
		"cost_center_id": "42",
		"env":            "production",
		"product_line":   "algorithms",
		"service":        "futura",
		"creator":        "jane.doe",
	}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != len(labels) {
		t.Fatalf("tag count = %d, want %d", len(tags), len(labels))
	}
	for k, v := range labels {
		got, ok := tags[k]
		if !ok {
			t.Errorf("missing tag %q", k)
			continue
		}
		if *got != v {
			t.Errorf("tag %q = %q, want %q", k, *got, v)
		}
	}
}

func TestEmrJobRunTags_NilLabels(t *testing.T) {
	emr := newTestEMREngine()
	if tags := emr.emrJobRunTags(nil, nil); tags != nil {
		t.Fatalf("expected nil for nil labels, got %v", tags)
	}
}

func TestEmrJobRunTags_EmptyLabels(t *testing.T) {
	emr := newTestEMREngine()
	if tags := emr.emrJobRunTags(state.Labels{}, nil); tags != nil {
		t.Fatalf("expected nil for empty labels, got %v", tags)
	}
}

func TestEmrJobRunTags_InvalidKeyDropped(t *testing.T) {
	emr := newTestEMREngine()
	labels := state.Labels{
		"valid_key":      "ok",
		"invalid\tkey":   "bad",
		"also invalid!":  "bad",
		"":               "empty-key",
	}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != 1 {
		t.Fatalf("tag count = %d, want 1", len(tags))
	}
	if _, ok := tags["valid_key"]; !ok {
		t.Error("expected valid_key to survive")
	}
}

func TestEmrJobRunTags_ValueTooLongDropped(t *testing.T) {
	emr := newTestEMREngine()
	longValue := make([]byte, 257)
	for i := range longValue {
		longValue[i] = 'a'
	}
	labels := state.Labels{
		"good": "ok",
		"bad":  string(longValue),
	}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != 1 {
		t.Fatalf("tag count = %d, want 1", len(tags))
	}
	if _, ok := tags["good"]; !ok {
		t.Error("expected 'good' tag to survive")
	}
}

func TestEmrJobRunTags_ExactlyMaxValueLength(t *testing.T) {
	emr := newTestEMREngine()
	exactValue := make([]byte, 256)
	for i := range exactValue {
		exactValue[i] = 'x'
	}
	labels := state.Labels{"key": string(exactValue)}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != 1 {
		t.Fatalf("tag count = %d, want 1", len(tags))
	}
}

func TestEmrJobRunTags_CapAt50(t *testing.T) {
	emr := newTestEMREngine()
	labels := make(state.Labels, 60)
	for i := 0; i < 60; i++ {
		labels[fmt.Sprintf("key_%03d", i)] = "val"
	}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != 50 {
		t.Fatalf("tag count = %d, want 50", len(tags))
	}
	// Sorted keys: key_000 through key_049 should survive
	for i := 0; i < 50; i++ {
		k := fmt.Sprintf("key_%03d", i)
		if _, ok := tags[k]; !ok {
			t.Errorf("expected %q to survive (deterministic sort)", k)
		}
	}
}

func TestEmrJobRunTags_AllInvalidReturnsNil(t *testing.T) {
	emr := newTestEMREngine()
	labels := state.Labels{
		"bad\x00key": "val",
		"also\nbad":  "val",
	}
	if tags := emr.emrJobRunTags(labels, nil); tags != nil {
		t.Fatalf("expected nil when all tags are invalid, got %v", tags)
	}
}

func TestEmrJobRunTags_SpecialCharsInKeyAllowed(t *testing.T) {
	emr := newTestEMREngine()
	labels := state.Labels{
		"cost/center":     "eng",
		"env.name":        "prod",
		"key with spaces": "ok",
		"a+b=c":           "math",
	}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != len(labels) {
		t.Fatalf("tag count = %d, want %d", len(tags), len(labels))
	}
}

func TestEmrJobRunTags_AWSPrefixDropped(t *testing.T) {
	emr := newTestEMREngine()
	labels := state.Labels{
		"aws:createdBy": "flotilla",
		"aws:foo":       "bar",
		"team":          "data-platform",
	}
	tags := emr.emrJobRunTags(labels, nil)
	if len(tags) != 1 {
		t.Fatalf("tag count = %d, want 1", len(tags))
	}
	if _, ok := tags["team"]; !ok {
		t.Error("expected 'team' to survive")
	}
}

func TestEmrJobRunTags_EnvFallbackWhenNoLabels(t *testing.T) {
	emr := newTestEMREngine()
	env := &state.EnvList{
		{Name: "team", Value: "portal"},
		{Name: "cost_center", Value: "eng"},
		{Name: "SOME_SECRET", Value: "should-not-appear"},
	}
	tags := emr.emrJobRunTags(nil, env)
	if len(tags) != 2 {
		t.Fatalf("tag count = %d, want 2", len(tags))
	}
	if *tags["team"] != "portal" {
		t.Errorf("team = %q, want %q", *tags["team"], "portal")
	}
	if *tags["cost_center"] != "eng" {
		t.Errorf("cost_center = %q, want %q", *tags["cost_center"], "eng")
	}
	if _, ok := tags["SOME_SECRET"]; ok {
		t.Error("SOME_SECRET should not be included in fallback tags")
	}
}

func TestEmrJobRunTags_EnvFallbackOnlyAllowedKeys(t *testing.T) {
	emr := newTestEMREngine()
	env := &state.EnvList{
		{Name: "team", Value: "data-platform"},
		{Name: "cost_center", Value: "CC-1234"},
		{Name: "env", Value: "production"},
		{Name: "product_line", Value: "algorithms"},
		{Name: "DATABASE_URL", Value: "postgres://secret"},
		{Name: "API_KEY", Value: "sk-12345"},
	}
	tags := emr.emrJobRunTags(nil, env)
	if len(tags) != 4 {
		t.Fatalf("tag count = %d, want 4", len(tags))
	}
	for _, k := range []string{"team", "cost_center", "env", "product_line"} {
		if _, ok := tags[k]; !ok {
			t.Errorf("expected %q in fallback tags", k)
		}
	}
	for _, k := range []string{"DATABASE_URL", "API_KEY"} {
		if _, ok := tags[k]; ok {
			t.Errorf("%q should not be in fallback tags", k)
		}
	}
}

func TestEmrJobRunTags_EnvFallbackSkipsNonAllowedKeys(t *testing.T) {
	emr := newTestEMREngine()
	env := &state.EnvList{
		{Name: "team", Value: "ok"},
		{Name: "random_key", Value: "skip"},
		{Name: "aws:internal", Value: "skip"},
	}
	tags := emr.emrJobRunTags(nil, env)
	if len(tags) != 1 {
		t.Fatalf("tag count = %d, want 1", len(tags))
	}
	if _, ok := tags["team"]; !ok {
		t.Error("expected team to survive")
	}
}

func TestEmrJobRunTags_LabelsPreferredOverEnv(t *testing.T) {
	emr := newTestEMREngine()
	labels := state.Labels{"team": "from-labels"}
	env := &state.EnvList{
		{Name: "team", Value: "from-env"},
		{Name: "extra", Value: "from-env"},
	}
	tags := emr.emrJobRunTags(labels, env)
	if len(tags) != 1 {
		t.Fatalf("tag count = %d, want 1 (labels only)", len(tags))
	}
	if *tags["team"] != "from-labels" {
		t.Errorf("team = %q, want %q", *tags["team"], "from-labels")
	}
}

func TestEMRContainersDefaultsConfSetsLoggingMemory(t *testing.T) {
	conf := emrContainersDefaultsConf()

	if got := *conf.Classification; got != emrContainersDefaultsClassification {
		t.Fatalf("classification = %q, want %q", got, emrContainersDefaultsClassification)
	}

	got, ok := conf.Properties[loggingRequestMemoryKey]
	if !ok {
		t.Fatalf("missing %q property", loggingRequestMemoryKey)
	}
	if *got != loggingRequestMemoryDefault {
		t.Fatalf("%s = %q, want %q", loggingRequestMemoryKey, *got, loggingRequestMemoryDefault)
	}
}

// env var name -> key within the client-credentials secret. DATAHUB_TOKEN is the
// one key that is upper-case.
var wantLakekeeperSecretKeys = map[string]string{
	"OAUTH2_CLIENT_ID":     "client_id",
	"OAUTH2_CLIENT_SECRET": "client_secret",
	"OAUTH2_SERVER_URI":    "token_url",
	"OAUTH2_SCOPE":         "scope",
	"CATALOG_URI":          "uri",
	"WAREHOUSE":            "warehouse",
	"DATAHUB_TOKEN":        "DATAHUB_TOKEN",
}

func TestLakekeeperSecretEnvVars(t *testing.T) {
	const secretName = "client-credentials"
	emr := &EMRExecutionEngine{lakekeeperSecretName: secretName}

	got := emr.lakekeeperSecretEnvVars()
	if len(got) != len(wantLakekeeperSecretKeys) {
		t.Fatalf("env var count = %d, want %d", len(got), len(wantLakekeeperSecretKeys))
	}

	seen := make(map[string]bool, len(got))
	for _, ev := range got {
		wantKey, ok := wantLakekeeperSecretKeys[ev.Name]
		if !ok {
			t.Errorf("unexpected env var %q", ev.Name)
			continue
		}
		seen[ev.Name] = true

		if ev.ValueFrom == nil || ev.ValueFrom.SecretKeyRef == nil {
			t.Errorf("%s: want a SecretKeyRef, got none", ev.Name)
			continue
		}
		ref := ev.ValueFrom.SecretKeyRef
		if ref.Name != secretName {
			t.Errorf("%s: secret = %q, want %q", ev.Name, ref.Name, secretName)
		}
		if ref.Key != wantKey {
			t.Errorf("%s: key = %q, want %q", ev.Name, ref.Key, wantKey)
		}
		// Optional keeps pods schedulable on clusters whose secret lacks the key.
		if ref.Optional == nil || !*ref.Optional {
			t.Errorf("%s: Optional not set to true", ev.Name)
		}
	}

	for name := range wantLakekeeperSecretKeys {
		if !seen[name] {
			t.Errorf("missing env var %q", name)
		}
	}
}

func TestLakekeeperSecretEnvVarsNotConfigured(t *testing.T) {
	emr := &EMRExecutionEngine{}
	if got := emr.lakekeeperSecretEnvVars(); got != nil {
		t.Fatalf("env var count = %d, want none when no secret is configured", len(got))
	}
}
