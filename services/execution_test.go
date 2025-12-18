package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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

func TestExecutionService_CommandHashCalculatedFromCommand(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that command_hash is MD5 of command, not description
	cmd := "python script.py --arg value"
	desc := "Different description"
	engine := state.DefaultEngine

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     &cmd,
			Description: &desc,
			OwnerID:     "testuser",
			Engine:      &engine,
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	if err != nil {
		t.Fatalf("Error creating run: %s", err.Error())
	}

	// Verify command_hash is MD5 of command
	expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(cmd)))
	if run.CommandHash == nil {
		t.Errorf("Expected non-nil command_hash")
	} else if *run.CommandHash != expectedHash {
		t.Errorf("Expected command_hash to be MD5 of command '%s', got '%s'", expectedHash, *run.CommandHash)
	}

	// Verify it's NOT MD5 of description
	descHash := fmt.Sprintf("%x", md5.Sum([]byte(desc)))
	if run.CommandHash != nil && *run.CommandHash == descHash {
		t.Errorf("command_hash should NOT be MD5 of description (that was the bug!)")
	}
}

func TestExecutionService_CommandHashWithSameDescriptionDifferentCommands(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that different commands get different hashes even with same description
	description := "Daily processing job"
	cmd1 := "python process.py --date 2025-01-01"
	cmd2 := "python process.py --date 2025-01-02"
	engine := state.DefaultEngine

	req1 := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     &cmd1,
			Description: &description,
			OwnerID:     "testuser",
			Engine:      &engine,
		},
	}

	req2 := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     &cmd2,
			Description: &description,
			OwnerID:     "testuser",
			Engine:      &engine,
		},
	}

	run1, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req1)
	if err != nil {
		t.Fatalf("Error creating run1: %s", err.Error())
	}

	run2, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req2)
	if err != nil {
		t.Fatalf("Error creating run2: %s", err.Error())
	}

	// Verify both have non-nil command_hash
	if run1.CommandHash == nil {
		t.Errorf("Expected run1 to have non-nil command_hash")
	}
	if run2.CommandHash == nil {
		t.Errorf("Expected run2 to have non-nil command_hash")
	}

	// Verify hashes are different (critical for ARA fix)
	if run1.CommandHash != nil && run2.CommandHash != nil {
		if *run1.CommandHash == *run2.CommandHash {
			t.Errorf("Different commands should have different hashes even with same description. "+
				"Both got hash '%s'. This was the ARA bug!", *run1.CommandHash)
		}
	}

	// Verify they match expected hashes
	expectedHash1 := fmt.Sprintf("%x", md5.Sum([]byte(cmd1)))
	expectedHash2 := fmt.Sprintf("%x", md5.Sum([]byte(cmd2)))

	if run1.CommandHash != nil && *run1.CommandHash != expectedHash1 {
		t.Errorf("run1 command_hash mismatch: expected '%s', got '%s'", expectedHash1, *run1.CommandHash)
	}
	if run2.CommandHash != nil && *run2.CommandHash != expectedHash2 {
		t.Errorf("run2 command_hash mismatch: expected '%s', got '%s'", expectedHash2, *run2.CommandHash)
	}
}

func TestExecutionService_CommandHashNullWhenCommandNull(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that NULL command results in NULL command_hash
	// (This is a malformed job, but should not crash)
	engine := state.DefaultEngine
	desc := "A description without a command"

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     nil, // NULL command
			Description: &desc,
			OwnerID:     "testuser",
			Engine:      &engine,
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	if err != nil {
		t.Fatalf("Error creating run: %s", err.Error())
	}

	// Command should be set from definition's command (if any)
	// But if definition also has no command, command_hash should be NULL
	if run.Command == nil || len(*run.Command) == 0 {
		// Command is NULL/empty, so command_hash should also be NULL
		if run.CommandHash != nil {
			t.Errorf("Expected NULL command_hash when command is NULL, got '%s'", *run.CommandHash)
		}
	}

	// Even if command gets set from definition, command_hash should NOT be from description
	if run.CommandHash != nil {
		descHash := fmt.Sprintf("%x", md5.Sum([]byte(desc)))
		if *run.CommandHash == descHash {
			t.Errorf("command_hash should NOT be MD5 of description (that was the bug!)")
		}
	}
}

func TestExecutionService_CommandHashMatchesCommand(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test with various command strings to ensure consistent hashing
	testCases := []struct {
		name    string
		command string
	}{
		{"Simple command", "echo hello"},
		{"Command with args", "python train.py --epochs 10 --lr 0.001"},
		{"Multi-line command", "set -e\necho 'Starting'\npython script.py\necho 'Done'"},
		{"Command with special chars", "grep -r 'pattern' /path/to/files | sort | uniq"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := state.DefaultEngine
			cmd := tc.command

			req := state.DefinitionExecutionRequest{
				ExecutionRequestCommon: &state.ExecutionRequestCommon{
					Command: &cmd,
					OwnerID: "testuser",
					Engine:  &engine,
				},
			}

			run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
			if err != nil {
				t.Fatalf("Error creating run: %s", err.Error())
			}

			expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(tc.command)))
			if run.CommandHash == nil {
				t.Errorf("Expected non-nil command_hash for command: %s", tc.command)
			} else if *run.CommandHash != expectedHash {
				t.Errorf("command_hash mismatch for '%s': expected '%s', got '%s'",
					tc.command, expectedHash, *run.CommandHash)
			}
		})
	}
}

func TestExecutionService_CommandHashStableAcrossRuns(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Verify same command always produces same hash (consistency check)
	cmd := "python train.py --model resnet50"
	engine := state.DefaultEngine

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command: &cmd,
			OwnerID: "testuser",
			Engine:  &engine,
		},
	}

	// Create multiple runs with same command
	run1, err1 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	run2, err2 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	run3, err3 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("Error creating runs")
	}

	// All should have same command_hash
	if run1.CommandHash == nil || run2.CommandHash == nil || run3.CommandHash == nil {
		t.Errorf("All runs should have non-nil command_hash")
	}

	if *run1.CommandHash != *run2.CommandHash || *run1.CommandHash != *run3.CommandHash {
		t.Errorf("Same command should always produce same hash. Got: '%s', '%s', '%s'",
			*run1.CommandHash, *run2.CommandHash, *run3.CommandHash)
	}

	// Verify it matches expected
	expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(cmd)))
	if *run1.CommandHash != expectedHash {
		t.Errorf("Expected hash '%s', got '%s'", expectedHash, *run1.CommandHash)
	}
}

func TestExecutionService_CommandHashNotSetInEndpoints(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that even if description is provided, command_hash comes from command
	// This verifies the endpoints.go fix (removal of description-based hashing)
	cmd := "python app.py"
	desc := "This is a description"
	engine := state.DefaultEngine

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     &cmd,
			Description: &desc,
			CommandHash: nil, // Explicitly NULL to verify it gets calculated
			OwnerID:     "testuser",
			Engine:      &engine,
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	if err != nil {
		t.Fatalf("Error creating run: %s", err.Error())
	}

	// Should be MD5 of command, not description
	cmdHash := fmt.Sprintf("%x", md5.Sum([]byte(cmd)))
	descHash := fmt.Sprintf("%x", md5.Sum([]byte(desc)))

	if run.CommandHash == nil {
		t.Errorf("Expected command_hash to be calculated")
	} else {
		if *run.CommandHash == descHash {
			t.Errorf("BUG: command_hash is MD5 of description! This should have been fixed.")
		}
		if *run.CommandHash != cmdHash {
			t.Errorf("Expected command_hash to be MD5 of command '%s', got '%s'", cmdHash, *run.CommandHash)
		}
	}
}

func TestExecutionService_CommandHashWithOverride(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that if API client explicitly provides a command_hash, it gets overwritten
	// by the correct hash calculated from the command
	cmd := "python script.py"
	wrongHash := "this_is_wrong_hash"
	engine := state.DefaultEngine

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     &cmd,
			CommandHash: aws.String(wrongHash), // Wrong hash provided by client
			OwnerID:     "testuser",
			Engine:      &engine,
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	if err != nil {
		t.Fatalf("Error creating run: %s", err.Error())
	}

	// Should be overwritten with correct hash
	expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(cmd)))
	if run.CommandHash == nil {
		t.Errorf("Expected non-nil command_hash")
	} else if *run.CommandHash == wrongHash {
		t.Errorf("BUG: Wrong hash was not overwritten! Still has '%s'", wrongHash)
	} else if *run.CommandHash != expectedHash {
		t.Errorf("Expected command_hash '%s', got '%s'", expectedHash, *run.CommandHash)
	}
}

func TestExecutionService_SparkCommandHashFromDescription(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that Spark jobs with NULL command get command_hash from description
	// Spark jobs don't have a command field - they store config in spark_extension
	desc := "Vmi Po Recon Data Extract / Run Snapshots"
	engine := state.EKSSparkEngine
	entryPoint := "s3://bucket/script.py"

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     nil, // Spark jobs have NULL command
			Description: &desc,
			OwnerID:     "testuser",
			Engine:      &engine,
			SparkExtension: &state.SparkExtension{
				SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
					EntryPoint: &entryPoint,
				},
			},
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	if err != nil {
		t.Fatalf("Error creating run: %s", err.Error())
	}

	// Should have command_hash from description (for Spark jobs)
	expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(desc)))
	if run.CommandHash == nil {
		t.Errorf("Expected non-nil command_hash for Spark job with description")
	} else if *run.CommandHash != expectedHash {
		t.Errorf("Expected Spark command_hash to be MD5 of description '%s', got '%s'", expectedHash, *run.CommandHash)
	}
}

func TestExecutionService_SparkCommandHashConsistent(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that Spark jobs with same description get same hash (critical for ARA)
	desc := "Vmi Po Recon Data Extract / Run Snapshots"
	engine := state.EKSSparkEngine
	entryPoint := "s3://bucket/script.py"

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     nil,
			Description: &desc,
			OwnerID:     "testuser",
			Engine:      &engine,
			SparkExtension: &state.SparkExtension{
				SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
					EntryPoint: &entryPoint,
				},
			},
		},
	}

	// Create multiple Spark runs with same description
	run1, err1 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	run2, err2 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	run3, err3 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("Error creating Spark runs")
	}

	// All should have same command_hash for ARA tracking
	if run1.CommandHash == nil || run2.CommandHash == nil || run3.CommandHash == nil {
		t.Errorf("All Spark runs should have non-nil command_hash")
	}

	if *run1.CommandHash != *run2.CommandHash || *run1.CommandHash != *run3.CommandHash {
		t.Errorf("Spark jobs with same description should always produce same hash. Got: '%s', '%s', '%s'",
			*run1.CommandHash, *run2.CommandHash, *run3.CommandHash)
	}

	// Verify it matches expected
	expectedHash := fmt.Sprintf("%x", md5.Sum([]byte(desc)))
	if *run1.CommandHash != expectedHash {
		t.Errorf("Expected Spark hash '%s', got '%s'", expectedHash, *run1.CommandHash)
	}
}

func TestExecutionService_SparkVsRegularEKSHashing(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that Spark and regular EKS jobs use different hashing strategies
	// This ensures no cross-contamination between Spark and regular jobs
	description := "Process data files"
	cmd := "python process.py"
	entryPoint := "s3://bucket/script.py"

	// Regular EKS job
	regularEngine := state.DefaultEngine
	regularReq := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     &cmd,
			Description: &description,
			OwnerID:     "testuser",
			Engine:      &regularEngine,
		},
	}

	// Spark job
	sparkEngine := state.EKSSparkEngine
	sparkReq := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     nil, // Spark has no command
			Description: &description,
			OwnerID:     "testuser",
			Engine:      &sparkEngine,
			SparkExtension: &state.SparkExtension{
				SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
					EntryPoint: &entryPoint,
				},
			},
		},
	}

	regularRun, err1 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &regularReq)
	sparkRun, err2 := es.CreateDefinitionRunByDefinitionID(ctx, "A", &sparkReq)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error creating runs")
	}

	// Verify both have command_hash
	if regularRun.CommandHash == nil {
		t.Errorf("Regular EKS job should have command_hash")
	}
	if sparkRun.CommandHash == nil {
		t.Errorf("Spark job should have command_hash")
	}

	// Verify they use different hash sources
	cmdHash := fmt.Sprintf("%x", md5.Sum([]byte(cmd)))
	descHash := fmt.Sprintf("%x", md5.Sum([]byte(description)))

	if regularRun.CommandHash != nil && *regularRun.CommandHash != cmdHash {
		t.Errorf("Regular EKS job should hash from command, expected '%s', got '%s'", cmdHash, *regularRun.CommandHash)
	}

	if sparkRun.CommandHash != nil && *sparkRun.CommandHash != descHash {
		t.Errorf("Spark job should hash from description, expected '%s', got '%s'", descHash, *sparkRun.CommandHash)
	}

	// Most importantly: they should have DIFFERENT hashes (no cross-contamination)
	if regularRun.CommandHash != nil && sparkRun.CommandHash != nil {
		if *regularRun.CommandHash == *sparkRun.CommandHash {
			t.Errorf("Regular EKS and Spark jobs should have different hashes to prevent ARA cross-contamination. Both got '%s'", *regularRun.CommandHash)
		}
	}
}

func TestExecutionService_SparkNullDescriptionNullHash(t *testing.T) {
	ctx := context.Background()
	es, _ := setUp(t)

	// Test that Spark jobs with NULL command AND NULL description get NULL hash
	// (This is a malformed job, but should not crash)
	engine := state.EKSSparkEngine
	entryPoint := "s3://bucket/script.py"

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Command:     nil, // Spark has no command
			Description: nil, // Also no description (malformed)
			OwnerID:     "testuser",
			Engine:      &engine,
			SparkExtension: &state.SparkExtension{
				SparkSubmitJobDriver: &state.SparkSubmitJobDriver{
					EntryPoint: &entryPoint,
				},
			},
		},
	}

	run, err := es.CreateDefinitionRunByDefinitionID(ctx, "A", &req)
	if err != nil {
		t.Fatalf("Error creating run: %s", err.Error())
	}

	// Should have NULL command_hash (malformed job)
	if run.CommandHash != nil {
		t.Errorf("Expected NULL command_hash for Spark job with NULL description, got '%s'", *run.CommandHash)
	}
}
