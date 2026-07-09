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
