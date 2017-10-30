package worker

import (
	"errors"
	gklog "github.com/go-kit/kit/log"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
	"os"
	"testing"
)

// Set up situation with runnable run
func setUpSubmitWorkerTest1(t *testing.T) (*submitWorker, *testutils.ImplementsAllTheThings) {
	validRun := state.Run{
		RunID:        "run:cupcake",
		DefinitionID: "def:cupcake",
		Status:       state.StatusQueued,
	}

	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	eventSinks := []flotillaLog.EventSink{flotillaLog.NewLocalEventSink()}
	logger := flotillaLog.NewLogger(l, eventSinks)

	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"def:cupcake": {DefinitionID: "def:cupcake"},
		},
		Runs: map[string]state.Run{
			"run:cupcake": validRun,
		},
		Qurls: map[string]string{
			"A": "a/",
		},
		Queued: []string{"run:cupcake"},
	}
	return &submitWorker{
		sm:  &imp,
		qm:  &imp,
		ee:  &imp,
		log: logger,
	}, &imp
}

// Set up situation with unrunnable run
func setUpSubmitWorkerTest2(t *testing.T) (*submitWorker, *testutils.ImplementsAllTheThings) {
	invalidRun := state.Run{
		RunID:        "run:shoebox",
		DefinitionID: "def:shoebox",
		Status:       state.StatusRunning,
	}

	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	eventSinks := []flotillaLog.EventSink{flotillaLog.NewLocalEventSink()}
	logger := flotillaLog.NewLogger(l, eventSinks)

	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"def:shoebox": {DefinitionID: "def:shoebox"},
		},
		Runs: map[string]state.Run{
			"run:shoebox": invalidRun,
		},
		Qurls: map[string]string{
			"A": "a/",
		},
		Queued: []string{"run:shoebox"},
	}
	return &submitWorker{
		sm:  &imp,
		qm:  &imp,
		ee:  &imp,
		log: logger,
	}, &imp
}

// Another unrunnable run
func setUpSubmitWorkerTest3(t *testing.T) (*submitWorker, *testutils.ImplementsAllTheThings) {
	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	eventSinks := []flotillaLog.EventSink{flotillaLog.NewLocalEventSink()}
	logger := flotillaLog.NewLogger(l, eventSinks)

	imp := testutils.ImplementsAllTheThings{
		T: t,
		Qurls: map[string]string{
			"A": "a/",
		},
		Queued: []string{"run:nope"},
	}
	return &submitWorker{
		sm:  &imp,
		qm:  &imp,
		ee:  &imp,
		log: logger,
	}, &imp
}

// we should only ack when
//   (a) run or def is missing
//   (b) status is not queued
//   (c) we hit a non-retryable error
//   (d) we successfully launch
// we should only NOT ack if
//   (a) we hit a retryable error

func TestSubmitWorker_Run(t *testing.T) {
	// 1. test that we only run queued runs
	// 2. test that we only run runs with a valid run and definition
	// 3. test that we don't ack on retryable errors, and properly set statusstopped on non-retryable errors

	// Test valid run; it's status is queued, it exists in state, its definition exists in state
	worker, imp := setUpSubmitWorkerTest1(t)
	worker.runOnce()

	expected := []string{"List", "Receive", "GetRun", "GetDefinition", "Execute", "UpdateRun", "RunReceipt.Done"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}
}

func TestSubmitWorker_Run2(t *testing.T) {
	// Test invalid run; it's status is running (this can happen with duplication in queues, which sqs allows)
	worker, imp := setUpSubmitWorkerTest2(t)
	worker.runOnce()

	// Importantly, execute is NOT called and it -is- acked
	expected := []string{"List", "Receive", "GetRun", "GetDefinition", "RunReceipt.Done"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}
}

func TestSubmitWorker_Run3(t *testing.T) {
	// Test invalid run; it's queued but does not exist; this should not happen
	// (run is queued but does not exist in state)
	worker, imp := setUpSubmitWorkerTest3(t)
	worker.runOnce()

	// Importantly, execute is NOT called and it -is- acked
	expected := []string{"List", "Receive", "GetRun", "RunReceipt.Done"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}
}

func TestSubmitWorker_Run4(t *testing.T) {
	// Test that we ack on non-retryable erorrs and change state to stopped
	worker, imp := setUpSubmitWorkerTest1(t)

	imp.ExecuteError = errors.New("nope")
	imp.ExecuteErrorIsRetryable = false

	worker.runOnce()

	// Importantly, execute is called and it -is- acked
	expected := []string{"List", "Receive", "GetRun", "GetDefinition", "Execute", "UpdateRun", "RunReceipt.Done"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}

	// Ensure the run gets updated to StatusQueued
	run, _ := imp.GetRun("run:cupcake")
	if run.Status != state.StatusStopped {
		t.Errorf("Expected submit worker to update run status to Stopped for non-retryable error")
	}
}

func TestSubmitWorker_Run5(t *testing.T) {
	// Test that we DON'T ack on retryable erorrs and don't change state
	worker, imp := setUpSubmitWorkerTest1(t)

	imp.ExecuteError = errors.New("nope")
	imp.ExecuteErrorIsRetryable = true

	worker.runOnce()

	// Importantly, execute it called but it is not updated nor is it acked
	expected := []string{"List", "Receive", "GetRun", "GetDefinition", "Execute"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	for i, call := range imp.Calls {
		if expected[i] != call {
			t.Errorf("Expected call %v to be %s but was %s", i, expected[i], call)
		}
	}
}
