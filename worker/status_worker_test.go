package worker

import (
	gklog "github.com/go-kit/kit/log"
	"github.com/stitchfix/flotilla-os/config"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
	"os"
	"testing"
)

func setUpStatusWorkerTest(t *testing.T) (*statusWorker, *testutils.ImplementsAllTheThings) {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)

	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	logger := flotillaLog.NewLogger(l, nil)
	run := state.Run{
		RunID:  "somerun",
		Status: state.StatusPending,
	}
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Qurls: map[string]string{
			"A": "a/",
		},
		Runs: map[string]state.Run{
			"somerun": run,
		},
		StatusUpdatesAsRuns: []state.Run{
			{
				TaskArn: "status1",
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "test",
					},
				},
				Status: state.StatusRunning,
			},
			{
				TaskArn: "status1",
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "test",
					},
				},
				Status: state.StatusPending,
			},
			{
				TaskArn: "status1",
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "test",
					},
				},
				Status: state.StatusStopped,
			},
		},
	}
	return &statusWorker{
		sm:   &imp,
		ee:   &imp,
		log:  logger,
		conf: c,
	}, &imp
}

func setUpStatusWorkerTest2(t *testing.T) (*statusWorker, *testutils.ImplementsAllTheThings) {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)

	l := gklog.NewLogfmtLogger(gklog.NewSyncWriter(os.Stderr))
	logger := flotillaLog.NewLogger(l, nil)
	run := state.Run{
		RunID:  "somerun",
		Status: state.StatusPending,
	}
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Qurls: map[string]string{
			"A": "a/",
		},
		Runs: map[string]state.Run{
			"somerun": run,
		},
		StatusUpdatesAsRuns: []state.Run{
			{
				TaskArn: "nope1",
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "prod",
					},
				},
				Status: state.StatusStopped,
			},
			{
				TaskArn: "nope2",
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "staging",
					},
				},
				Status: state.StatusStopped,
			},
			{
				TaskArn: "status1",
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "test",
					},
				},
				Status: state.StatusRunning,
			},
		},
	}
	return &statusWorker{
		sm:   &imp,
		ee:   &imp,
		log:  logger,
		conf: c,
	}, &imp
}
func TestStatusWorker_Run(t *testing.T) {
	// 1. With a valid status update, run should go from PENDING to RUNNING
	// 2. With a valid status update that is out of order, run status stays the same
	//    eg. RUNNING does *not* transition back to PENDING
	worker, imp := setUpStatusWorkerTest(t)

	worker.runOnce()

	expected := []string{"PollStatus", "ListRuns", "UpdateRun", "StatusReceipt.Done"}
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	run, _ := imp.GetRun("somerun")
	if run.Status != state.StatusRunning {
		t.Errorf("Expected run to have updated status: %s but was %s", state.StatusRunning, run.Status)
	}

	worker.runOnce()
	run, _ = imp.GetRun("somerun")
	if run.Status != state.StatusRunning {
		t.Errorf("Expected run to have same status: %s, but was %s", state.StatusRunning, run.Status)
	}

	worker.runOnce()
	run, _ = imp.GetRun("somerun")
	if run.Status != state.StatusStopped {
		t.Errorf("Expected run to have updated status: %s, but was %s", state.StatusStopped, run.Status)
	}
}

func TestStatusWorker_Run2(t *testing.T) {
	//
	// Ignore and ack status updates that don't belong to us
	//
	worker, imp := setUpStatusWorkerTest2(t)

	//
	// The first iterations correspond to the first two mock status updates which
	// don't belong to the test mode and should be ignored and acked
	//
	expected := []string{"ReceiveStatus", "StatusReceipt.Done"}
	worker.runOnce()

	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	imp.Calls = []string{}
	worker.runOnce()
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	imp.Calls = []string{}
	expected = []string{"ReceiveStatus", "ListRuns", "UpdateRun", "StatusReceipt.Done"}
	worker.runOnce()
	if len(imp.Calls) != len(expected) {
		t.Errorf("Unexpected number of run calls, expected %v but was %v", len(expected), len(imp.Calls))
	}

	run, _ := imp.GetRun("somerun")
	if run.Status != state.StatusRunning {
		t.Errorf("Expected run to have updated status: %s but was %s", state.StatusRunning, run.Status)
	}
}
