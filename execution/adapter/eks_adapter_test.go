package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

func TestRoundCPUMillicores(t *testing.T) {
	adapter := &eksAdapter{}

	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		// The problematic case that triggered this fix
		{"1024m rounds to 1000m", 1024, 1000},

		// Edge cases around quarters
		{"1000m stays 1000m", 1000, 1000},
		{"1125m rounds to 1250m", 1125, 1250},
		{"1150m rounds to 1250m", 1150, 1250},
		{"1250m stays 1250m", 1250, 1250},

		// Test rounding up and down
		{"100m rounds to 0m", 100, 0},
		{"125m rounds to 250m", 125, 250},
		{"137m rounds to 250m", 137, 250},
		{"250m stays 250m", 250, 250},
		{"374m rounds to 250m", 374, 250},
		{"375m rounds to 500m", 375, 500},
		{"500m stays 500m", 500, 500},
		{"624m rounds to 500m", 624, 500},
		{"625m rounds to 750m", 625, 750},
		{"750m stays 750m", 750, 750},

		// Higher values - test both rounding up and down
		{"2048m rounds to 2000m", 2048, 2000},
		{"2100m rounds to 2000m", 2100, 2000},
		{"2126m rounds UP to 2250m", 2126, 2250},
		{"3000m stays 3000m", 3000, 3000},
		{"3001m rounds to 3000m", 3001, 3000},
		{"3126m rounds UP to 3250m", 3126, 3250},
		{"3200m rounds UP to 3250m", 3200, 3250},

		// Large values
		{"60000m stays 60000m", 60000, 60000},
		{"60024m rounds to 60000m", 60024, 60000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.roundCPUMillicores(tt.input)
			if result != tt.expected {
				t.Errorf("roundCPUMillicores(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRoundCPUAvoidsCgroupIssue verifies that rounded values avoid the systemd
// cgroup rounding issue where non-integer percentages get rounded up by systemd
func TestRoundCPUAvoidsCgroupIssue(t *testing.T) {
	adapter := &eksAdapter{}

	// Test values that would cause systemd rounding issues
	problematicValues := []int64{
		1024, // 102.4% -> systemd rounds to 103%
		1025, // 102.5% -> systemd rounds to 103%
		1026, // 102.6% -> systemd rounds to 103%
		2048, // 204.8% -> systemd rounds to 205%
		3072, // 307.2% -> systemd rounds to 308%
	}

	for _, input := range problematicValues {
		result := adapter.roundCPUMillicores(input)

		// Verify result is a multiple of 250 (quarter core)
		if result%250 != 0 {
			t.Errorf("roundCPUMillicores(%d) = %d, which is not a multiple of 250m", input, result)
		}

		// Verify result produces an integer percentage (whole or quarter)
		// Valid: 0%, 25%, 50%, 75%, 100%, 125%, etc.
		// 1000m = 100%, 250m = 25%
		percentage := (result * 100) / 1000 // percentage with 1 decimal place
		if percentage%25 != 0 {
			t.Errorf("roundCPUMillicores(%d) = %d, which produces non-quarter percentage (%d)",
				input, result, percentage)
		}
	}
}

// mockLogger implements flotillaLog.Logger for testing
type mockLogger struct {
	logCalls   [][]interface{}
	eventCalls [][]interface{}
}

func (m *mockLogger) Log(keyvals ...interface{}) error {
	m.logCalls = append(m.logCalls, keyvals)
	return nil
}

func (m *mockLogger) Event(keyvals ...interface{}) error {
	m.eventCalls = append(m.eventCalls, keyvals)
	return nil
}

func (m *mockLogger) reset() {
	m.logCalls = nil
	m.eventCalls = nil
}

// mockStateManager implements state.Manager for testing
type mockStateManager struct {
	estimateResourcesResult state.TaskResources
	estimateResourcesError  error
}

func (m *mockStateManager) EstimateRunResources(ctx context.Context, executableID string, commandHash string) (state.TaskResources, error) {
	return m.estimateResourcesResult, m.estimateResourcesError
}

// Stub implementations for required interface methods
func (m *mockStateManager) Name() string                      { return "mock" }
func (m *mockStateManager) Initialize(conf config.Config) error { return nil }
func (m *mockStateManager) Cleanup() error                                    { return nil }
func (m *mockStateManager) ListDefinitions(ctx context.Context, limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string) (state.DefinitionList, error) {
	return state.DefinitionList{}, nil
}
func (m *mockStateManager) GetDefinition(ctx context.Context, definitionID string) (state.Definition, error) {
	return state.Definition{}, nil
}
func (m *mockStateManager) GetDefinitionByAlias(ctx context.Context, alias string) (state.Definition, error) {
	return state.Definition{}, nil
}
func (m *mockStateManager) UpdateDefinition(ctx context.Context, definitionID string, updates state.Definition) (state.Definition, error) {
	return state.Definition{}, nil
}
func (m *mockStateManager) CreateDefinition(ctx context.Context, d state.Definition) error { return nil }
func (m *mockStateManager) DeleteDefinition(ctx context.Context, definitionID string) error { return nil }
func (m *mockStateManager) ListRuns(ctx context.Context, limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string, engines []string) (state.RunList, error) {
	return state.RunList{}, nil
}
func (m *mockStateManager) EstimateExecutorCount(ctx context.Context, executableID string, commandHash string) (int64, error) {
	return 0, nil
}
func (m *mockStateManager) ExecutorOOM(ctx context.Context, executableID string, commandHash string) (bool, error) {
	return false, nil
}
func (m *mockStateManager) DriverOOM(ctx context.Context, executableID string, commandHash string) (bool, error) {
	return false, nil
}
func (m *mockStateManager) GetRun(ctx context.Context, runID string) (state.Run, error) {
	return state.Run{}, nil
}
func (m *mockStateManager) CreateRun(ctx context.Context, r state.Run) error { return nil }
func (m *mockStateManager) UpdateRun(ctx context.Context, runID string, updates state.Run) (state.Run, error) {
	return state.Run{}, nil
}
func (m *mockStateManager) ListGroups(ctx context.Context, limit int, offset int, name *string) (state.GroupsList, error) {
	return state.GroupsList{}, nil
}
func (m *mockStateManager) ListTags(ctx context.Context, limit int, offset int, name *string) (state.TagsList, error) {
	return state.TagsList{}, nil
}
func (m *mockStateManager) ListWorkers(ctx context.Context, engine string) (state.WorkersList, error) {
	return state.WorkersList{}, nil
}
func (m *mockStateManager) BatchUpdateWorkers(ctx context.Context, updates []state.Worker) (state.WorkersList, error) {
	return state.WorkersList{}, nil
}
func (m *mockStateManager) GetWorker(ctx context.Context, workerType string, engine string) (state.Worker, error) {
	return state.Worker{}, nil
}
func (m *mockStateManager) UpdateWorker(ctx context.Context, workerType string, updates state.Worker) (state.Worker, error) {
	return state.Worker{}, nil
}
func (m *mockStateManager) GetExecutableByTypeAndID(ctx context.Context, executableType state.ExecutableType, executableID string) (state.Executable, error) {
	return state.Definition{}, nil
}
func (m *mockStateManager) GetTemplateByID(ctx context.Context, templateID string) (state.Template, error) {
	return state.Template{}, nil
}
func (m *mockStateManager) GetLatestTemplateByTemplateName(ctx context.Context, templateName string) (bool, state.Template, error) {
	return false, state.Template{}, nil
}
func (m *mockStateManager) GetTemplateByVersion(ctx context.Context, templateName string, templateVersion int64) (bool, state.Template, error) {
	return false, state.Template{}, nil
}
func (m *mockStateManager) ListTemplates(ctx context.Context, limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	return state.TemplateList{}, nil
}
func (m *mockStateManager) ListTemplatesLatestOnly(ctx context.Context, limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	return state.TemplateList{}, nil
}
func (m *mockStateManager) CreateTemplate(ctx context.Context, t state.Template) error { return nil }
func (m *mockStateManager) ListFailingNodes(ctx context.Context) (state.NodeList, error) {
	return state.NodeList{}, nil
}
func (m *mockStateManager) GetPodReAttemptRate(ctx context.Context) (float32, error) {
	return 0, nil
}
func (m *mockStateManager) GetNodeLifecycle(ctx context.Context, executableID string, commandHash string) (string, error) {
	return "", nil
}
func (m *mockStateManager) GetTaskHistoricalRuntime(ctx context.Context, executableID string, runId string) (float32, error) {
	return 0, nil
}
func (m *mockStateManager) CheckIdempotenceKey(ctx context.Context, idempotenceKey string) (string, error) {
	return "", nil
}
func (m *mockStateManager) GetRunByEMRJobId(ctx context.Context, emrJobId string) (state.Run, error) {
	return state.Run{}, nil
}
func (m *mockStateManager) GetResources(ctx context.Context, runID string) (state.Run, error) {
	return state.Run{}, nil
}
func (m *mockStateManager) ListClusterStates(ctx context.Context) ([]state.ClusterMetadata, error) {
	return nil, nil
}
func (m *mockStateManager) UpdateClusterMetadata(ctx context.Context, cluster state.ClusterMetadata) error {
	return nil
}
func (m *mockStateManager) DeleteClusterMetadata(ctx context.Context, clusterID string) error {
	return nil
}
func (m *mockStateManager) GetClusterByID(ctx context.Context, clusterID string) (state.ClusterMetadata, error) {
	return state.ClusterMetadata{}, nil
}
func (m *mockStateManager) GetRunStatus(ctx context.Context, runID string) (state.RunStatus, error) {
	return state.RunStatus{}, nil
}

// mockExecutable implements state.Executable for testing
type mockExecutable struct {
	executableID string
	resources    *state.ExecutableResources
}

func (m *mockExecutable) GetExecutableID() *string {
	return &m.executableID
}

func (m *mockExecutable) GetExecutableType() *state.ExecutableType {
	t := state.ExecutableTypeDefinition
	return &t
}

func (m *mockExecutable) GetExecutableResources() *state.ExecutableResources {
	return m.resources
}

func (m *mockExecutable) GetExecutableCommand(req state.ExecutionRequest) (string, error) {
	return "", nil
}

func (m *mockExecutable) GetExecutableResourceName() string {
	return m.executableID
}

func TestAdaptiveResources_NonGPUJob_ARAEnabled_Success(t *testing.T) {
	logger := &mockLogger{}
	adapter, err := NewEKSAdapter(logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	executableID := "test-executable"
	executable := &mockExecutable{
		executableID: executableID,
		resources: &state.ExecutableResources{
			Memory: int64Ptr(1000),
			Cpu:    int64Ptr(500),
		},
	}

	run := state.Run{
		RunID:        "test-run",
		ExecutableID: &executableID,
	}

	manager := &mockStateManager{
		estimateResourcesResult: state.TaskResources{
			Cpu:    2000,
			Memory: 3000,
		},
		estimateResourcesError: nil,
	}

	// Note: We can't easily test metrics emission since they're package-level functions,
	// but we can verify the logic works correctly
	cpuLimit, memLimit, cpuRequest, memRequest := adapter.(*eksAdapter).adaptiveResources(
		context.Background(),
		executable,
		run,
		manager,
		true, // araEnabled
	)

	// Verify ARA increased resources
	if cpuRequest != 2000 {
		t.Errorf("Expected CPU request 2000, got %d", cpuRequest)
	}
	if memRequest != 3000 {
		t.Errorf("Expected memory request 3000, got %d", memRequest)
	}
	if cpuLimit != 2000 {
		t.Errorf("Expected CPU limit 2000, got %d", cpuLimit)
	}
	if memLimit != 3000 {
		t.Errorf("Expected memory limit 3000, got %d", memLimit)
	}
}

func TestAdaptiveResources_GPUJob_SkipsARA(t *testing.T) {
	logger := &mockLogger{}
	adapter, err := NewEKSAdapter(logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	executableID := "test-executable"
	gpu := int64(1)
	executable := &mockExecutable{
		executableID: executableID,
		resources: &state.ExecutableResources{
			Memory: int64Ptr(1000),
			Cpu:    int64Ptr(500),
		},
	}

	run := state.Run{
		RunID:        "test-run",
		ExecutableID: &executableID,
		Gpu:          &gpu,
	}

	manager := &mockStateManager{}

	_, _, cpuRequest, memRequest := adapter.(*eksAdapter).adaptiveResources(
		context.Background(),
		executable,
		run,
		manager,
		true, // araEnabled
	)

	// Verify GPU jobs use defaults (no ARA)
	defaultCPU := int64(500)
	defaultMem := int64(1000)
	if cpuRequest != defaultCPU {
		t.Errorf("Expected CPU request %d (default), got %d", defaultCPU, cpuRequest)
	}
	if memRequest != defaultMem {
		t.Errorf("Expected memory request %d (default), got %d", defaultMem, memRequest)
	}
}

func TestAdaptiveResources_EstimationFailed(t *testing.T) {
	logger := &mockLogger{}
	adapter, err := NewEKSAdapter(logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	executableID := "test-executable"
	executable := &mockExecutable{
		executableID: executableID,
		resources: &state.ExecutableResources{
			Memory: int64Ptr(1000),
			Cpu:    int64Ptr(500),
		},
	}

	run := state.Run{
		RunID:        "test-run",
		ExecutableID: &executableID,
	}

	manager := &mockStateManager{
		estimateResourcesError: errors.New("estimation failed"),
	}

	_, _, cpuRequest, memRequest := adapter.(*eksAdapter).adaptiveResources(
		context.Background(),
		executable,
		run,
		manager,
		true, // araEnabled
	)

	// Verify defaults are used when estimation fails
	defaultCPU := int64(500)
	defaultMem := int64(1000)
	if cpuRequest != defaultCPU {
		t.Errorf("Expected CPU request %d (default), got %d", defaultCPU, cpuRequest)
	}
	if memRequest != defaultMem {
		t.Errorf("Expected memory request %d (default), got %d", defaultMem, memRequest)
	}
}

func TestAdaptiveResources_MaxResourceBoundsHit(t *testing.T) {
	logger := &mockLogger{}
	adapter, err := NewEKSAdapter(logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	executableID := "test-executable"
	definitionID := "test-definition"
	command := "test-command"
	executable := &mockExecutable{
		executableID: executableID,
		resources: &state.ExecutableResources{
			Memory: int64Ptr(1000),
			Cpu:    int64Ptr(500),
		},
	}

	run := state.Run{
		RunID:        "test-run",
		ExecutableID: &executableID,
		DefinitionID: definitionID,
		Command:      &command,
		ClusterName:  "test-cluster",
	}

	// Return resources that exceed max bounds
	manager := &mockStateManager{
		estimateResourcesResult: state.TaskResources{
			Cpu:    state.MaxCPU + 10000, // Exceeds max
			Memory: state.MaxMem + 50000, // Exceeds max
		},
		estimateResourcesError: nil,
	}

	cpuLimit, memLimit, cpuRequest, memRequest := adapter.(*eksAdapter).adaptiveResources(
		context.Background(),
		executable,
		run,
		manager,
		true, // araEnabled
	)

	// Verify resources are capped at max bounds
	if cpuRequest != state.MaxCPU {
		t.Errorf("Expected CPU request capped at %d, got %d", state.MaxCPU, cpuRequest)
	}
	if memRequest != state.MaxMem {
		t.Errorf("Expected memory request capped at %d, got %d", state.MaxMem, memRequest)
	}
	if cpuLimit != state.MaxCPU {
		t.Errorf("Expected CPU limit capped at %d, got %d", state.MaxCPU, cpuLimit)
	}
	if memLimit != state.MaxMem {
		t.Errorf("Expected memory limit capped at %d, got %d", state.MaxMem, memLimit)
	}

	// Verify logger was called for max resource hit
	// There should be two logs: one for ARA adjustment, one for max bounds hit
	if len(logger.logCalls) < 2 {
		t.Errorf("Expected at least 2 logger.Log calls (ARA adjustment + max bounds hit), got %d", len(logger.logCalls))
		return
	}
	// Find the max bounds hit log (should have level:warn)
	var maxBoundsLog []interface{}
	for _, logCall := range logger.logCalls {
		for i := 0; i < len(logCall); i += 2 {
			if i+1 < len(logCall) && logCall[i] == "level" && logCall[i+1] == "warn" {
				maxBoundsLog = logCall
				break
			}
		}
		if maxBoundsLog != nil {
			break
		}
	}
	if maxBoundsLog == nil {
		t.Errorf("Expected log with level:warn for max bounds hit, got logCalls: %v", logger.logCalls)
		return
	}
	// Verify log contains expected fields
	foundMessage := false
	foundRunID := false
	for i := 0; i < len(maxBoundsLog); i += 2 {
		if i+1 < len(maxBoundsLog) {
			key := maxBoundsLog[i]
			value := maxBoundsLog[i+1]
			if key == "message" {
				msg := value.(string)
				if msg == "ARA resource allocation hit maximum limit" || msg == "ARA memory allocation hit maximum limit - potential over-provisioning" {
					foundMessage = true
				}
			}
			if key == "run_id" && value == "test-run" {
				foundRunID = true
			}
		}
	}
	if !foundMessage {
		t.Errorf("Expected log to contain message about max resource hit")
	}
	if !foundRunID {
		t.Error("Expected log to contain 'run_id: test-run'")
	}
}

func TestAdaptiveResources_ARADisabled(t *testing.T) {
	logger := &mockLogger{}
	adapter, err := NewEKSAdapter(logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	executableID := "test-executable"
	executable := &mockExecutable{
		executableID: executableID,
		resources: &state.ExecutableResources{
			Memory: int64Ptr(1000),
			Cpu:    int64Ptr(500),
		},
	}

	run := state.Run{
		RunID:        "test-run",
		ExecutableID: &executableID,
	}

	manager := &mockStateManager{}

	_, _, cpuRequest, memRequest := adapter.(*eksAdapter).adaptiveResources(
		context.Background(),
		executable,
		run,
		manager,
		false, // araEnabled = false
	)

	// Verify defaults are used when ARA is disabled
	defaultCPU := int64(500)
	defaultMem := int64(1000)
	if cpuRequest != defaultCPU {
		t.Errorf("Expected CPU request %d (default), got %d", defaultCPU, cpuRequest)
	}
	if memRequest != defaultMem {
		t.Errorf("Expected memory request %d (default), got %d", defaultMem, memRequest)
	}
}

func TestEmitARAMetrics_StructuredLog(t *testing.T) {
	logger := &mockLogger{}
	adapter, err := NewEKSAdapter(logger)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	executableID := "test-executable"
	definitionID := "test-definition"
	command := "test-command"
	run := state.Run{
		RunID:        "test-run",
		ExecutableID: &executableID,
		DefinitionID: definitionID,
		Command:      &command,
		ClusterName:  "test-cluster",
	}

	adapter.(*eksAdapter).emitARAMetrics(run, 1000, 2000, 3000, 4000, 5000, 6000, true, true)

	// Verify logger was called
	if len(logger.logCalls) == 0 {
		t.Error("Expected logger.Log to be called")
		return
	}

	logCall := logger.logCalls[0]
	expectedFields := map[string]interface{}{
		"level":                  "warn",
		"message":                "ARA memory allocation hit maximum limit - potential over-provisioning",
		"run_id":                 "test-run",
		"cluster":                 "test-cluster",
		"default_cpu_millicores": int64(1000),
		"default_memory_mb":       int64(2000),
		"requested_cpu_millicores": int64(5000),
		"requested_memory_mb":       int64(6000),
		"final_cpu_millicores":     int64(3000),
		"final_memory_mb":          int64(4000),
		"max_cpu_hit":             true,
		"max_memory_hit":           true,
		"definition_id":           "test-definition",
		"executable_id":           "test-executable",
		"command":                 "test-command",
		"memory_overage_mb":       int64(2000), // 6000 - 4000
		"cpu_overage_millicores":  int64(2000), // 5000 - 3000
	}

	// Verify all expected fields are present
	logMap := make(map[interface{}]interface{})
	for i := 0; i < len(logCall); i += 2 {
		if i+1 < len(logCall) {
			logMap[logCall[i]] = logCall[i+1]
		}
	}

	for key, expectedValue := range expectedFields {
		if actualValue, ok := logMap[key]; !ok {
			t.Errorf("Expected log to contain field '%s'", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected log field '%s' to be %v, got %v", key, expectedValue, actualValue)
		}
	}
}

func TestEmitARAMetrics_NilLogger(t *testing.T) {
	// Create adapter with nil logger (shouldn't panic)
	adapter := &eksAdapter{logger: nil}

	run := state.Run{
		RunID: "test-run",
	}

	// Should not panic
	adapter.emitARAMetrics(run, 1000, 2000, 3000, 4000, 5000, 6000, true, true)
}

// Helper function
func int64Ptr(i int64) *int64 {
	return &i
}
