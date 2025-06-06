package services

import (
	"context"
	"log"
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
		ClusterStates: []state.ClusterMetadata{
			{Name: "cluster1", Status: state.StatusActive, StatusReason: "Active and healthy"},
			{Name: "cluster2", Status: state.StatusActive, StatusReason: "Active and healthy"},
		},
	}

	es, err := NewExecutionService(c, &imp, &imp, &imp, &imp)
	if err != nil {
		log.Fatalf("error seting up execution service: %s", err.Error())
	}
	return es, &imp
}

func TestExecutionService_CreateDefinitionRunByDefinitionID(t *testing.T) {
	ctx := context.Background()
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
		"ListClusterStates":        true,
	}

	cmd := "_test_cmd_"
	sa := "fooAccount"
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
			ServiceAccount:   &sa,
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "B", &req)
	if err != nil {
		t.Error(err.Error())
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

	if run.ServiceAccount == nil {
		t.Errorf("Expected non-nil service account")
	} else {
		if *run.ServiceAccount != sa {
			t.Errorf("Unexpected service account, found [%s], exptecting [%s]", *run.ServiceAccount, sa)
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
	ctx := context.Background()
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
		"ListClusterStates":        true,
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
	run, err := es.CreateDefinitionRunByAlias(ctx, "aliasB", &req)
	if err != nil {
		t.Error(err.Error())
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
	ctx := context.Background()
	es, imp := setUp(t)
	es.List(ctx, 1, 0, "asc", "cluster_name", nil, nil)

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
	ctx := context.Background()
	es, imp := setUp(t)
	es.List(
		ctx, 1, 0,
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
func TestExecutionService_ListClusters(t *testing.T) {
	ctx := context.Background()
	es, imp := setUp(t)

	clusters, err := es.ListClusters(ctx)
	if err != nil {
		t.Errorf("Expected no error listing clusters, got: %v", err)
	}

	expectedCalls := map[string]bool{
		"ListClusterStates": true,
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during cluster listing: %s", call)
		}
	}

	if len(clusters) != 2 {
		t.Errorf("Expected 2 clusters, got %d", len(clusters))
	}
}

func TestExecutionService_CreateDefinitionRunWithTier(t *testing.T) {
	ctx := context.Background()
	// Set up test environment
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)

	// Create mock implementation with clusters supporting different tiers
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"A": {DefinitionID: "A", Alias: "aliasA"},
		},
		Runs: map[string]state.Run{},
		Qurls: map[string]string{
			"A": "a/",
		},
		ClusterStates: []state.ClusterMetadata{
			{
				Name:         "prod-cluster",
				Status:       state.StatusActive,
				StatusReason: "Active and healthy",
				AllowedTiers: []string{"1", "2"},
			},
			{
				Name:         "staging-cluster",
				Status:       state.StatusActive,
				StatusReason: "Active and healthy",
				AllowedTiers: []string{"3", "4"},
			},
			{
				Name:         "string-cluster",
				Status:       state.StatusActive,
				StatusReason: "Active and healthy",
				AllowedTiers: []string{"tier3", "tier4"},
			},
			{
				Name:         "unrestricted-cluster",
				Status:       state.StatusActive,
				StatusReason: "Active and healthy",
				// No tiers specified - should use default tier
			},
			{
				Name:         "maintenance-cluster",
				Status:       state.StatusMaintenance,
				StatusReason: "In maintenance",
				AllowedTiers: []string{"1", "2", "3", "4"},
			},
		},
	}

	imp.GetRandomClusterName = func(clusters []string) string {
		if len(clusters) > 0 {
			return clusters[0]
		}
		return ""
	}

	es, err := NewExecutionService(c, &imp, &imp, &imp, &imp)
	if err != nil {
		t.Fatalf("Error setting up execution service: %s", err.Error())
	}

	// Test cases with different tiers
	testCases := []struct {
		name            string
		tier            string
		expectedCluster string
	}{
		{
			name:            "Production tier request",
			tier:            "1",
			expectedCluster: "prod-cluster",
		},
		{
			name:            "Staging tier request",
			tier:            "3",
			expectedCluster: "staging-cluster",
		},
		{
			name:            "No tier specified",
			tier:            "",
			expectedCluster: "staging-cluster",
		},
		{
			name:            "String Tier",
			tier:            "tier3",
			expectedCluster: "string-cluster",
		},
		{
			name:            "Invalid tier",
			tier:            "nonexistent",
			expectedCluster: es.GetDefaultCluster(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imp.Calls = make([]string, 0)
			cmd := "echo test"
			engine := state.DefaultEngine
			req := state.DefinitionExecutionRequest{
				ExecutionRequestCommon: &state.ExecutionRequestCommon{
					Tier:    state.Tier(tc.tier),
					Command: &cmd,
					OwnerID: "testuser",
					Engine:  &engine,
				},
			}

			run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
			if err != nil {
				t.Errorf("Error creating run: %s", err.Error())
				return
			}
			// Verify the selected cluster matches expectations
			if run.ClusterName != tc.expectedCluster {
				t.Errorf("Expected cluster %s for tier %s, but got %s",
					tc.expectedCluster, tc.tier, run.ClusterName)
			}

			// Verify tier was set correctly
			if string(run.Tier) != tc.tier && tc.tier != "" {
				t.Errorf("Expected tier %s, but got %s", tc.tier, string(run.Tier))
			}
		})
	}
}

func TestExecutionService_GetRunStatus(t *testing.T) {
	ctx := context.Background()
	es, imp := setUp(t)

	expectedCalls := map[string]bool{
		"GetRunStatus": true,
	}

	status, err := es.GetRunStatus(ctx, "runA")

	if err != nil {
		t.Errorf("Expected no error when getting status of existing run, got: %s", err.Error())
	}

	if len(imp.Calls) != len(expectedCalls) {
		t.Errorf("Expected exactly %v calls during status retrieval but was: %v", len(expectedCalls), len(imp.Calls))
	}

	for _, call := range imp.Calls {
		_, ok := expectedCalls[call]
		if !ok {
			t.Errorf("Unexpected call during status retrieval: %s", call)
		}
	}

	if status.RunID != "runA" {
		t.Errorf("Expected run ID 'runA' but got '%s'", status.RunID)
	}

	if status.DefinitionID != "A" {
		t.Errorf("Expected definition ID 'A' but got '%s'", status.DefinitionID)
	}

	if status.ClusterName != "A" {
		t.Errorf("Expected cluster name 'A' but got '%s'", status.ClusterName)
	}

	imp.Calls = []string{}

	_, err = es.GetRunStatus(ctx, "nonexistent")

	if err == nil {
		t.Errorf("Expected error when getting status of non-existent run, got nil")
	}

	expectedErrorString := "No run with ID: nonexistent"
	if err != nil && err.Error() != expectedErrorString {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorString, err.Error())
	}

}
