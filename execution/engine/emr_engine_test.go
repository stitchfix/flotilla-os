package engine

import "testing"

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
