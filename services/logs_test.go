package services

import (
	"testing"

	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
)

func setUpLogServiceTest(t *testing.T) (LogService, *testutils.ImplementsAllTheThings) {
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"B": {DefinitionID: "{}"},
		},
		Runs: map[string]state.Run{
			"isQueued": {DefinitionID: "q", RunID: "isQueued", Status: state.StatusQueued},
			"running":  {DefinitionID: "B", RunID: "running", Status: state.StatusRunning},
		},
	}
	ls, _ := NewLogService(&imp, &imp)
	return ls, &imp
}

func TestLogService_Logs(t *testing.T) {
	ls, imp := setUpLogServiceTest(t)

	//
	// Check that we don't try to get logs for runs that won't have them yet
	//

	expectedCalls := map[string]bool{
		"GetRun": true,
	}

	_, _, err := ls.Logs("isQueued", nil, nil, nil)
	if err != nil {
		t.Error(err.Error())
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls for log query for queued run but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during log query for queued run: %s", call)
		}
	}

	//
	// Check that we do get logs for runs that should have them
	//
	ls, imp = setUpLogServiceTest(t)
	expectedCalls = map[string]bool{
		"GetRun":                   true,
		"GetDefinition":            true,
		"Logs":                     true,
		"GetExecutableByTypeAndID": true,
	}

	_, _, err = ls.Logs("running", nil, nil, nil)
	if err != nil {
		t.Error(err.Error())
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls for log query for running run but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during log query for running run: %s", call)
		}
	}
}
