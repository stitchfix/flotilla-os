package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/state"
)

// mockLogger implements log.Logger for testing
type mockLogger struct {
	logCalls   [][]interface{}
	eventCalls [][]interface{}
}

// Compile-time check to ensure mockLogger implements log.Logger
var _ log.Logger = (*mockLogger)(nil)

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
		ClusterName:  "test-cluster",
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

	// Verify logger was called with ARA adjustment message
	if len(logger.logCalls) < 1 {
		t.Error("Expected logger.Log to be called for ARA adjustment")
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
		ClusterName:  "test-cluster",
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
		ClusterName:  "test-cluster",
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

	// Verify logger was called for max resource hits (should have multiple log calls)
	if len(logger.logCalls) < 2 {
		t.Errorf("Expected at least 2 log calls (ARA adjustment + max limit warnings), got %d", len(logger.logCalls))
	}

	// Verify one of the log calls contains a warning about hitting max limits
	foundWarning := false
	for _, logCall := range logger.logCalls {
		for i := 0; i < len(logCall); i += 2 {
			if i+1 < len(logCall) {
				key := logCall[i]
				value := logCall[i+1]
				if key == "level" && value == "warn" {
					foundWarning = true
					break
				}
			}
		}
		if foundWarning {
			break
		}
	}

	if !foundWarning {
		t.Error("Expected at least one warning log for hitting max resource limits")
	}
}

// Helper function
func int64Ptr(i int64) *int64 {
	return &i
}
