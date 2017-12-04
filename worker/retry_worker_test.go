package worker

import (
	gklog "github.com/go-kit/kit/log"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
	"os"
	"testing"
)

func setUpRetryWorkerTest(t *testing.T) (*retryWorker, *testutils.ImplementsAllTheThings) {
	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	logger := flotillaLog.NewLogger(l, nil)

	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"A": {DefinitionID: "A"},
			"B": {DefinitionID: "B"},
			"C": {DefinitionID: "C", Image: "invalidimage"},
		},
		Runs: map[string]state.Run{
			"runA": {DefinitionID: "A", ClusterName: "A", GroupName: "A", RunID: "runA"},
		},
		Qurls: map[string]string{
			"A": "a/",
			"B": "b/",
		},
	}
	return &retryWorker{
		sm:  &imp,
		ee:  &imp,
		log: logger,
	}, &imp
}

func TestRetryWorker_Run(t *testing.T) {
	worker, imp := setUpRetryWorkerTest(t)
	worker.runOnce()

	//
	// Make sure that the worker resets the status to StatusQueued, and calls the appropriate methods
	// in order (get runs to retry, get qurls for them, update them to queued status, then enqueue them)
	//
	expected := []string{"ListRuns", "UpdateRun", "Enqueue"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}

	// Ensure the run gets updated to StatusQueued
	run, _ := imp.GetRun("runA")
	if run.Status != state.StatusQueued {
		t.Errorf("Expected retry worker to update run status to Queued")
	}
}
