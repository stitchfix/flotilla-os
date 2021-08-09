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
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "test",
					},
				},
				Status: state.StatusRunning,
			},
			{
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "test",
					},
				},
				Status: state.StatusPending,
			},
			{
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
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "prod",
					},
				},
				Status: state.StatusStopped,
			},
			{
				Env: &state.EnvList{
					{
						Name:  "FLOTILLA_SERVER_MODE",
						Value: "staging",
					},
				},
				Status: state.StatusStopped,
			},
			{
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
