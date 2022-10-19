package services

import (
	"testing"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
)

func setUp(t *testing.T) (ExecutionService, *testutils.ImplementsAllTheThings) {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"A": {DefinitionID: "A", Alias: "aliasA"},
			"B": {DefinitionID: "B", Alias: "aliasB"},
			"C": {DefinitionID: "C", Alias: "aliasC", ExecutableResources: state.ExecutableResources{Image: "invalidimage"}},
		},
		Runs: map[string]state.Run{
			"runA": {DefinitionID: "A", ClusterName: "A", GroupName: "A", RunID: "runA"},
			"runB": {DefinitionID: "B", ClusterName: "B", GroupName: "B", RunID: "runB"},
		},
		Qurls: map[string]string{
			"A": "a/",
			"B": "b/",
		},
	}
	es, _ := NewExecutionService(c, &imp, &imp, &imp, &imp)
	return es, &imp
}

func TestExecutionService_CreateDefinitionRunByDefinitionID(t *testing.T) {
	// Tests valid create
	es, imp := setUp(t)
	env := &state.EnvList{
		{Name: "K1", Value: "V1"},
	}
	expectedCalls := map[string]bool{
		"GetDefinition":            true,
		"CreateRun":                true,
		"UpdateRun":                true,
		"GetTaskHistoricalRuntime": true,
		"GetPodReAttemptRate":      true,
		"Enqueue":                  true,
	}

	cmd := "_test_cmd_"
	cpu := int64(512)
	engine := state.DefaultEngine
	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			ClusterName:      "clusta",
			Env:              env,
			OwnerID:          "somebody",
			Command:          &cmd,
			Memory:           nil,
			Cpu:              &cpu,
			Engine:           &engine,
			EphemeralStorage: nil,
			NodeLifecycle:    nil,
			IdempotenceKey:   nil,
			Arch:             nil,
		},
	}
	run, err := es.CreateDefinitionRunByDefinitionID("B", &req)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls during run creation but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during run creation: %s", call)
		}
	}

	if len(run.RunID) == 0 {
		t.Errorf("Expected Create to populated run with non-empty RunID")
	}

	if run.DefinitionID != "B" {
		t.Errorf("Expected definitionID 'B' but was '%s'", run.DefinitionID)
	}

	if run.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status '%s' but was '%s'", state.StatusQueued, run.Status)
	}

	if run.User != "somebody" {
		t.Errorf("Expected new run to have user 'somebody' but was '%s'", run.User)
	}

	if run.QueuedAt == nil {
		t.Errorf("Expected new run to have a 'queued_at' field but was nil.")
	}

	if run.Env == nil {
		t.Errorf("Expected non-nil environment")
	}

	if len(*run.Env) != (len(es.ReservedVariables()) + len(*env)) {
		t.Errorf("Unexpected number of environment variables; expected %v but was %v",
			len(es.ReservedVariables())+len(*env), len(*run.Env))
	}

	if run.Command == nil {
		t.Errorf("Expected non-nil command")
	} else {
		if *run.Command != cmd {
			t.Errorf("Unexpected command, found [%s], exptecting [%s]", *run.Command, cmd)
		}
	}

	if run.Cpu == nil {
		t.Errorf("Expected non-nil cpu")
	} else {
		if *run.Cpu != cpu {
			t.Errorf("Unexpected cpu, found [%d], exptecting [%d]", *run.Cpu, cpu)
		}
	}

	includesExpected := false
	for _, e := range *run.Env {
		if e.Name == "K1" && e.Value == "V1" {
			includesExpected = true
		}
	}

	if !includesExpected {
		t.Errorf("Expected K1:V1 in run environment")
	}
}

func TestExecutionService_CreateDefinitionRunByAlias(t *testing.T) {
	// Tests valid create
	es, imp := setUp(t)
	env := &state.EnvList{
		{Name: "K1", Value: "V1"},
	}
	expectedCalls := map[string]bool{
		"GetDefinitionByAlias":     true,
		"CreateRun":                true,
		"UpdateRun":                true,
		"GetTaskHistoricalRuntime": true,
		"GetPodReAttemptRate":      true,
		"Enqueue":                  true,
	}
	mem := int64(1024)
	engine := state.DefaultEngine
	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			ClusterName:      "",
			Env:              env,
			OwnerID:          "somebody",
			Command:          nil,
			Memory:           &mem,
			Cpu:              nil,
			Engine:           &engine,
			EphemeralStorage: nil,
			NodeLifecycle:    nil,
			IdempotenceKey:   nil,
			Arch:             nil,
		},
	}
	run, err := es.CreateDefinitionRunByAlias("aliasB", &req)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls during run creation but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during run creation: %s", call)
		}
	}

	if len(run.RunID) == 0 {
		t.Errorf("Expected Create to populated run with non-empty RunID")
	}

	if run.DefinitionID != "B" {
		t.Errorf("Expected definitionID 'B' but was '%s'", run.DefinitionID)
	}

	if run.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status '%s' but was '%s'", state.StatusQueued, run.Status)
	}

	if run.User != "somebody" {
		t.Errorf("Expected new run to have user 'somebody' but was '%s'", run.User)
	}

	if run.QueuedAt == nil {
		t.Errorf("Expected new run to have a 'queued_at' field but was nil.")
	}

	if run.Env == nil {
		t.Errorf("Expected non-nil environment")
	}

	if len(*run.Env) != (len(es.ReservedVariables()) + len(*env)) {
		t.Errorf("Unexpected number of environment variables; expected %v but was %v",
			len(es.ReservedVariables())+len(*env), len(*run.Env))
	}

	if run.Memory == nil {
		t.Errorf("Expected non-nil memory")
	} else {
		if *run.Memory != mem {
			t.Errorf("Unexpected memory , found [%d], exptecting [%d]", *run.Memory, mem)
		}
	}

	includesExpected := false
	for _, e := range *run.Env {
		if e.Name == "K1" && e.Value == "V1" {
			includesExpected = true
		}
	}

	if !includesExpected {
		t.Errorf("Expected K1:V1 in run environment")
	}
}

func TestExecutionService_List(t *testing.T) {
	es, imp := setUp(t)
	es.List(1, 0, "asc", "cluster_name", nil, nil)

	expectedCalls := map[string]bool{
		"ListRuns": true,
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls during run list with no filters but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during run list with no filters: %s", call)
		}
	}
}

func TestExecutionService_List2(t *testing.T) {
	es, imp := setUp(t)
	es.List(
		1, 0,
		"asc", "cluster_name",
		map[string][]string{"definition_id": {"A"}}, nil)

	expectedCalls := map[string]bool{
		"GetDefinition": true,
		"ListRuns":      true,
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls during run list with no filters but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during run list with no filters: %s", call)
		}
	}
}
