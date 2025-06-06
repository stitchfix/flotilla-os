package testutils

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/execution/engine"
	"github.com/stitchfix/flotilla-os/queue"
	"github.com/stitchfix/flotilla-os/state"
)

// ImplementsAllTheThings defines a struct which implements many of the interfaces
// to facilitate easier testing
type ImplementsAllTheThings struct {
	T                       *testing.T
	Calls                   []string                    // Collects calls
	Definitions             map[string]state.Definition // Definitions stored in "state"
	Runs                    map[string]state.Run        // Runs stored in "state"
	Workers                 []state.Worker              // Workers stored in "state"
	Qurls                   map[string]string           // Urls returned by Queue Manager
	Defined                 []string                    // List of defined definitions (Execution Engine)
	Queued                  []string                    // List of queued runs (Queue Manager)
	StatusUpdates           []string                    // List of queued status updates (Queue Manager)
	StatusUpdatesAsRuns     []state.Run                 // List of queued status updates (Execution Engine)
	ExecuteError            error                       // Execution Engine - error to return
	ExecuteErrorIsRetryable bool                        // Execution Engine - is the run retryable?
	Groups                  []string
	Tags                    []string
	Templates               map[string]state.Template
	ClusterStates           []state.ClusterMetadata
	GetRandomClusterName    func(clusters []string) string
}

func (iatt *ImplementsAllTheThings) GetResources(ctx context.Context, runID string) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "GetResources")
	run, exists := iatt.Runs[runID]
	if !exists {
		return state.Run{}, fmt.Errorf("Run with id %s not found", runID)
	}
	return run, nil
}

func (iatt *ImplementsAllTheThings) ListClusters() ([]state.ClusterMetadata, error) {
	iatt.Calls = append(iatt.Calls, "ListClusters")
	return iatt.ClusterStates, nil
}

func (i *ImplementsAllTheThings) ListClusterStates(ctx context.Context) ([]state.ClusterMetadata, error) {
	i.Calls = append(i.Calls, "ListClusterStates")
	fmt.Printf("ListClusterStates called, returning %d clusters\n", len(i.ClusterStates))
	return i.ClusterStates, nil
}

func (i *ImplementsAllTheThings) GetClusterByID(ctx context.Context, clusterID string) (state.ClusterMetadata, error) {
	i.Calls = append(i.Calls, "GetClusterByID")
	return i.ClusterStates[0], nil
}

func (i *ImplementsAllTheThings) DeleteClusterMetadata(ctx context.Context, clusterName string) error {
	i.Calls = append(i.Calls, "DeleteClusterMetadata")
	return nil
}

func (i *ImplementsAllTheThings) UpdateClusterMetadata(ctx context.Context, cluster state.ClusterMetadata) error {
	i.Calls = append(i.Calls, "UpdateClusterMetadata")
	return nil
}

func (iatt *ImplementsAllTheThings) LogsText(executable state.Executable, run state.Run, w http.ResponseWriter) error {
	iatt.Calls = append(iatt.Calls, "LogsText")
	return nil
}

func (iatt *ImplementsAllTheThings) Log(keyvals ...interface{}) error {
	iatt.Calls = append(iatt.Calls, "Name")
	return nil
}

func (iatt *ImplementsAllTheThings) Event(keyvals ...interface{}) error {
	iatt.Calls = append(iatt.Calls, "Name")
	return nil
}

// Name - general
func (iatt *ImplementsAllTheThings) Name() string {
	iatt.Calls = append(iatt.Calls, "Name")
	return "implementer"
}

// Initialize - general
func (iatt *ImplementsAllTheThings) Initialize(conf config.Config) error {
	iatt.Calls = append(iatt.Calls, "Initialize")
	return nil
}

// Cleanup - general
func (iatt *ImplementsAllTheThings) Cleanup() error {
	iatt.Calls = append(iatt.Calls, "Cleanup")
	return nil
}

func (iatt *ImplementsAllTheThings) ListFailingNodes(ctx context.Context) (state.NodeList, error) {
	var nodeList state.NodeList
	iatt.Calls = append(iatt.Calls, "ListFailingNodes")
	return nodeList, nil
}

func (iatt *ImplementsAllTheThings) GetPodReAttemptRate(ctx context.Context) (float32, error) {
	iatt.Calls = append(iatt.Calls, "GetPodReAttemptRate")
	return 1.0, nil
}

func (iatt *ImplementsAllTheThings) GetNodeLifecycle(ctx context.Context, executableID string, commandHash string) (string, error) {
	iatt.Calls = append(iatt.Calls, "GetNodeLifecycle")
	return "spot", nil
}

func (iatt *ImplementsAllTheThings) GetTaskHistoricalRuntime(ctx context.Context, executableID string, runId string) (float32, error) {
	iatt.Calls = append(iatt.Calls, "GetTaskHistoricalRuntime")
	return 1.0, nil
}

// ListDefinitions - StateManager
func (iatt *ImplementsAllTheThings) ListDefinitions(
	ctx context.Context,
	limit int, offset int, sortBy string,
	order string, filters map[string][]string,
	envFilters map[string]string) (state.DefinitionList, error) {
	iatt.Calls = append(iatt.Calls, "ListDefinitions")
	dl := state.DefinitionList{Total: len(iatt.Definitions)}
	for _, d := range iatt.Definitions {
		dl.Definitions = append(dl.Definitions, d)
	}
	return dl, nil
}

// GetDefinition - StateManager
func (iatt *ImplementsAllTheThings) GetDefinition(ctx context.Context, definitionID string) (state.Definition, error) {
	iatt.Calls = append(iatt.Calls, "GetDefinition")
	var err error
	d, ok := iatt.Definitions[definitionID]
	if !ok {
		err = fmt.Errorf("No definition %s", definitionID)
	}
	return d, err
}

// GetDefinitionByAlias - StateManager
func (iatt *ImplementsAllTheThings) GetDefinitionByAlias(ctx context.Context, alias string) (state.Definition, error) {
	iatt.Calls = append(iatt.Calls, "GetDefinitionByAlias")
	for _, d := range iatt.Definitions {
		if d.Alias == alias {
			return d, nil
		}
	}
	return state.Definition{}, fmt.Errorf("No definition with alias %s", alias)
}

// UpdateDefinition - StateManager
func (iatt *ImplementsAllTheThings) UpdateDefinition(ctx context.Context, definitionID string, updates state.Definition) (state.Definition, error) {
	iatt.Calls = append(iatt.Calls, "UpdateDefinition")
	defn := iatt.Definitions[definitionID]
	defn.UpdateWith(updates)
	iatt.Definitions[definitionID] = defn
	return defn, nil
}

// CreateDefinition - StateManager
func (iatt *ImplementsAllTheThings) CreateDefinition(ctx context.Context, d state.Definition) error {
	iatt.Calls = append(iatt.Calls, "CreateDefinition")
	iatt.Definitions[d.DefinitionID] = d
	return nil
}

// DeleteDefinition - StateManager
func (iatt *ImplementsAllTheThings) DeleteDefinition(ctx context.Context, definitionID string) error {
	iatt.Calls = append(iatt.Calls, "DeleteDefinition")
	delete(iatt.Definitions, definitionID)
	return nil
}

// ListRuns - StateManager
func (iatt *ImplementsAllTheThings) ListRuns(ctx context.Context, limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string, engines []string) (state.RunList, error) {
	iatt.Calls = append(iatt.Calls, "ListRuns")
	rl := state.RunList{Total: len(iatt.Runs)}
	for _, r := range iatt.Runs {
		rl.Runs = append(rl.Runs, r)
	}
	return rl, nil
}

// GetRun - StateManager
func (iatt *ImplementsAllTheThings) GetRun(ctx context.Context, runID string) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "GetRun")
	var err error
	r, ok := iatt.Runs[runID]
	if !ok {
		err = fmt.Errorf("No run %s", runID)
	}
	return r, err
}

func (iatt *ImplementsAllTheThings) GetRunByEMRJobId(ctx context.Context, emrJobId string) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "GetRunByEMRJobId")
	var err error
	r, ok := iatt.Runs[emrJobId]
	if !ok {
		err = fmt.Errorf("No run %s", emrJobId)
	}
	return r, err
}

// CreateRun - StateManager
func (iatt *ImplementsAllTheThings) CreateRun(ctx context.Context, r state.Run) error {
	iatt.Calls = append(iatt.Calls, "CreateRun")
	iatt.Runs[r.RunID] = r
	return nil
}

func (iatt *ImplementsAllTheThings) EstimateRunResources(ctx context.Context, executableID string, command string) (state.TaskResources, error) {
	iatt.Calls = append(iatt.Calls, "EstimateRunResources")
	return state.TaskResources{}, nil
}

func (iatt *ImplementsAllTheThings) EstimateExecutorCount(ctx context.Context, executableID string, commandHash string) (int64, error) {
	iatt.Calls = append(iatt.Calls, "EstimateExecutorCount")
	return 0, nil
}

func (iatt *ImplementsAllTheThings) ExecutorOOM(ctx context.Context, executableID string, commandHash string) (bool, error) {
	iatt.Calls = append(iatt.Calls, "ExecutorOOM")
	return false, nil
}
func (iatt *ImplementsAllTheThings) DriverOOM(ctx context.Context, executableID string, commandHash string) (bool, error) {
	iatt.Calls = append(iatt.Calls, "DriverOOM")
	return false, nil
}

// UpdateRun - StateManager
func (iatt *ImplementsAllTheThings) UpdateRun(ctx context.Context, runID string, updates state.Run) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "UpdateRun")
	run := iatt.Runs[runID]
	run.UpdateWith(updates)
	iatt.Runs[runID] = run
	return run, nil
}

// ListGroups - StateManager
func (iatt *ImplementsAllTheThings) ListGroups(ctx context.Context, limit int, offset int, name *string) (state.GroupsList, error) {
	iatt.Calls = append(iatt.Calls, "ListGroups")
	return state.GroupsList{Total: len(iatt.Groups), Groups: iatt.Groups}, nil
}

// ListTags - StateManager
func (iatt *ImplementsAllTheThings) ListTags(ctx context.Context, limit int, offset int, name *string) (state.TagsList, error) {
	iatt.Calls = append(iatt.Calls, "ListTags")
	return state.TagsList{Total: len(iatt.Tags), Tags: iatt.Tags}, nil
}

// initWorkerTable - StateManager
func (iatt *ImplementsAllTheThings) initWorkerTable(c config.Config) error {
	iatt.Calls = append(iatt.Calls, "initWorkerTable")
	return nil
}

// ListWorkers - StateManager
func (iatt *ImplementsAllTheThings) ListWorkers(ctx context.Context, engine string) (state.WorkersList, error) {
	iatt.Calls = append(iatt.Calls, "ListWorkers")
	return state.WorkersList{Total: len(iatt.Workers), Workers: iatt.Workers}, nil
}

func (iatt *ImplementsAllTheThings) CheckIdempotenceKey(ctx context.Context, idempotenceKey string) (string, error) {
	iatt.Calls = append(iatt.Calls, "CheckIdempotenceKey")
	return "42", nil
}

// GetWorker - StateManager
func (iatt *ImplementsAllTheThings) GetWorker(ctx context.Context, workerType string, engine string) (state.Worker, error) {
	iatt.Calls = append(iatt.Calls, "GetWorker")
	return state.Worker{WorkerType: workerType, CountPerInstance: 2}, nil
}

// UpdateWorker - StateManager
func (iatt *ImplementsAllTheThings) UpdateWorker(ctx context.Context, workerType string, updates state.Worker) (state.Worker, error) {
	iatt.Calls = append(iatt.Calls, "UpdateWorker")
	return state.Worker{WorkerType: workerType, CountPerInstance: updates.CountPerInstance}, nil
}

// BatchUpdateWorkers- StateManager
func (iatt *ImplementsAllTheThings) BatchUpdateWorkers(ctx context.Context, updates []state.Worker) (state.WorkersList, error) {
	iatt.Calls = append(iatt.Calls, "BatchUpdateWorkers")
	return state.WorkersList{Total: len(iatt.Workers), Workers: iatt.Workers}, nil
}

// QurlFor - QueueManager
func (iatt *ImplementsAllTheThings) QurlFor(name string, prefixed bool) (string, error) {
	iatt.Calls = append(iatt.Calls, "QurlFor")
	qurl, _ := iatt.Qurls[name]
	return qurl, nil
}

func (iatt *ImplementsAllTheThings) Enqueue(ctx context.Context, run state.Run) error {
	iatt.Calls = append(iatt.Calls, "Enqueue")
	iatt.Queued = append(iatt.Queued, run.RunID)
	return nil
}

// ReceiveRun - QueueManager
func (iatt *ImplementsAllTheThings) ReceiveRun(qURL string) (queue.RunReceipt, error) {
	iatt.Calls = append(iatt.Calls, "ReceiveRun")
	if len(iatt.Queued) == 0 {
		return queue.RunReceipt{}, nil
	}

	popped := iatt.Queued[0]
	iatt.Queued = iatt.Queued[1:]
	receipt := queue.RunReceipt{
		Run: &state.Run{RunID: popped},
	}
	receipt.Done = func() error {
		iatt.Calls = append(iatt.Calls, "RunReceipt.Done")
		return nil
	}
	return receipt, nil
}

// ReceiveStatus - QueueManager
func (iatt *ImplementsAllTheThings) ReceiveStatus(qURL string) (queue.StatusReceipt, error) {
	iatt.Calls = append(iatt.Calls, "ReceiveStatus")
	if len(iatt.StatusUpdates) == 0 {
		return queue.StatusReceipt{}, nil
	}

	popped := iatt.StatusUpdates[0]
	iatt.StatusUpdates = iatt.StatusUpdates[1:]

	receipt := queue.StatusReceipt{
		StatusUpdate: &popped,
	}

	receipt.Done = func() error {
		iatt.Calls = append(iatt.Calls, "RunReceipt.Done")
		return nil
	}
	return receipt, nil
}

// List - QueueManager
func (iatt *ImplementsAllTheThings) List() ([]string, error) {
	iatt.Calls = append(iatt.Calls, "List")
	res := make([]string, len(iatt.Qurls))
	i := 0
	for _, qurl := range iatt.Qurls {
		res[i] = qurl
		i++
	}
	return res, nil
}

func (iatt *ImplementsAllTheThings) GetEvents(ctx context.Context, run state.Run) (state.PodEventList, error) {
	iatt.Calls = append(iatt.Calls, "GetEvents")

	return state.PodEventList{
		Total:     0,
		PodEvents: nil,
	}, nil
}

func (iatt *ImplementsAllTheThings) FetchUpdateStatus(ctx context.Context, run state.Run) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "FetchUpdateStatus")

	return run, nil
}

func (iatt *ImplementsAllTheThings) FetchPodMetrics(ctx context.Context, run state.Run) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "FetchPodMetrics")
	return run, nil
}

// CanBeRun - Cluster Client
func (iatt *ImplementsAllTheThings) CanBeRun(clusterName string, executableResources state.ExecutableResources) (bool, error) {
	iatt.Calls = append(iatt.Calls, "CanBeRun")
	if clusterName == "invalidcluster" {
		return false, nil
	}
	return true, nil
}

// IsImageValid - Registry Client
func (iatt *ImplementsAllTheThings) IsImageValid(imageRef string) (bool, error) {
	iatt.Calls = append(iatt.Calls, "IsImageValid")
	if imageRef == "invalidimage" {
		return false, nil
	}
	return true, nil
}

func (iatt *ImplementsAllTheThings) PollRunStatus(ctx context.Context) (state.Run, error) {
	iatt.Calls = append(iatt.Calls, "PollRunStatus")
	return state.Run{}, nil
}

// PollRuns - Execution Engine
func (iatt *ImplementsAllTheThings) PollRuns(ctx context.Context) ([]engine.RunReceipt, error) {
	iatt.Calls = append(iatt.Calls, "PollRuns")

	var r []engine.RunReceipt
	if len(iatt.Queued) == 0 {
		return r, nil
	}

	popped := iatt.Queued[0]
	iatt.Queued = iatt.Queued[1:]
	receipt := queue.RunReceipt{
		Run: &state.Run{RunID: popped},
	}
	receipt.Done = func() error {
		iatt.Calls = append(iatt.Calls, "RunReceipt.Done")
		return nil
	}
	r = append(r, engine.RunReceipt{receipt, 1111, 1111111, 1})
	return r, nil
}

// PollStatus - Execution Engine
func (iatt *ImplementsAllTheThings) PollStatus(ctx context.Context) (engine.RunReceipt, error) {
	iatt.Calls = append(iatt.Calls, "PollStatus")
	if len(iatt.StatusUpdatesAsRuns) == 0 {
		return engine.RunReceipt{}, nil
	}

	popped := iatt.StatusUpdatesAsRuns[0]
	iatt.StatusUpdatesAsRuns = iatt.StatusUpdatesAsRuns[1:]

	receipt := queue.RunReceipt{
		Run: &popped,
	}

	receipt.Done = func() error {
		iatt.Calls = append(iatt.Calls, "StatusReceipt.Done")
		return nil
	}
	return engine.RunReceipt{receipt, 1111, 1111111, 1}, nil
}

// Execute - Execution Engine
func (iatt *ImplementsAllTheThings) Execute(ctx context.Context, executable state.Executable, run state.Run, manager state.Manager) (state.Run, bool, error) {
	iatt.Calls = append(iatt.Calls, "Execute")
	return state.Run{}, iatt.ExecuteErrorIsRetryable, iatt.ExecuteError
}

// Terminate - Execution Engine
func (iatt *ImplementsAllTheThings) Terminate(ctx context.Context, run state.Run) error {
	iatt.Calls = append(iatt.Calls, "Terminate")
	return nil
}

// Define - Execution Engine
func (iatt *ImplementsAllTheThings) Define(ctx context.Context, definition state.Definition) (state.Definition, error) {
	iatt.Calls = append(iatt.Calls, "Define")
	iatt.Defined = append(iatt.Defined, definition.DefinitionID)
	return definition, nil
}

// Deregister - Execution Engine
func (iatt *ImplementsAllTheThings) Deregister(ctx context.Context, definition state.Definition) error {
	iatt.Calls = append(iatt.Calls, "Deregister")
	return nil
}

// Logs - Logs Client
func (iatt *ImplementsAllTheThings) Logs(executable state.Executable, run state.Run, lastSeen *string, role *string, facility *string) (string, *string, error) {
	iatt.Calls = append(iatt.Calls, "Logs")
	return "", aws.String(""), nil
}

// GetExecutableByTypeAndID - StateManager
func (iatt *ImplementsAllTheThings) GetExecutableByTypeAndID(ctx context.Context, t state.ExecutableType, id string) (state.Executable, error) {
	iatt.Calls = append(iatt.Calls, "GetExecutableByTypeAndID")
	switch t {
	case state.ExecutableTypeDefinition:
		return iatt.GetDefinition(ctx, id)
	case state.ExecutableTypeTemplate:
		return iatt.GetTemplateByID(ctx, id)
	default:
		return nil, fmt.Errorf("Invalid executable type %s", t)
	}
}

// ListTemplates - StateManager
func (iatt *ImplementsAllTheThings) ListTemplates(ctx context.Context, limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	iatt.Calls = append(iatt.Calls, "ListTemplates")
	tl := state.TemplateList{Total: len(iatt.Templates)}
	for _, t := range iatt.Templates {
		tl.Templates = append(tl.Templates, t)
	}
	return tl, nil
}

// ListTemplatesLatestOnly - StateManager
func (iatt *ImplementsAllTheThings) ListTemplatesLatestOnly(ctx context.Context, limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	// TODO: this is not actually implemented correctly - but also we're never
	// using it.
	iatt.Calls = append(iatt.Calls, "ListTemplatesLatestOnly")
	tl := state.TemplateList{Total: len(iatt.Templates)}
	for _, t := range iatt.Templates {
		tl.Templates = append(tl.Templates, t)
	}
	return tl, nil
}

func (iatt *ImplementsAllTheThings) GetTemplateByVersion(ctx context.Context, templateName string, templateVersion int64) (bool, state.Template, error) {
	iatt.Calls = append(iatt.Calls, "GetTemplateByVersion")
	var err error
	var tpl *state.Template

	// Iterate over templates to find max version.
	for _, t := range iatt.Templates {
		if t.TemplateName == templateName && t.Version == templateVersion {
			tpl = &t
		}
	}

	if tpl == nil {
		return false, *tpl, fmt.Errorf("No template with name: %s", templateName)
	}

	return true, *tpl, err
}

// GetTemplateByID - StateManager
func (iatt *ImplementsAllTheThings) GetTemplateByID(ctx context.Context, id string) (state.Template, error) {
	iatt.Calls = append(iatt.Calls, "GetTemplateByID")
	var err error
	t, ok := iatt.Templates[id]
	if !ok {
		err = fmt.Errorf("No template %s", id)
	}
	return t, err
}

// GetLatestTemplateByTemplateName - StateManager
func (iatt *ImplementsAllTheThings) GetLatestTemplateByTemplateName(ctx context.Context, templateName string) (bool, state.Template, error) {
	iatt.Calls = append(iatt.Calls, "GetLatestTemplateByTemplateName")
	var err error
	var tpl *state.Template
	var maxVersion int64 = int64(math.Inf(-1))

	// Iterate over templates to find max version.
	for _, t := range iatt.Templates {
		if t.TemplateName == templateName && t.Version > maxVersion {
			tpl = &t
			maxVersion = t.Version
		}
	}

	if tpl == nil {
		return false, *tpl, fmt.Errorf("No template with name: %s", templateName)
	}

	return true, *tpl, err
}

// CreateTemplate - StateManager
func (iatt *ImplementsAllTheThings) CreateTemplate(ctx context.Context, t state.Template) error {
	iatt.Calls = append(iatt.Calls, "CreateTemplate")
	iatt.Templates[t.TemplateID] = t
	return nil
}

func (iatt *ImplementsAllTheThings) GetRunStatus(ctx context.Context, runID string) (state.RunStatus, error) {
	iatt.Calls = append(iatt.Calls, "GetRunStatus")
	var err error

	r, ok := iatt.Runs[runID]
	if !ok {
		err = fmt.Errorf("No run with ID: %s", runID)
		return state.RunStatus{}, err
	}

	status := state.RunStatus{
		RunID:        r.RunID,
		Status:       r.Status,
		DefinitionID: r.DefinitionID,
		ClusterName:  r.ClusterName,
		QueuedAt:     r.QueuedAt,
		StartedAt:    r.StartedAt,
		FinishedAt:   r.FinishedAt,
		ExitCode:     r.ExitCode,
		ExitReason:   r.ExitReason,
		Engine:       r.Engine,
		Alias:        r.Alias,
	}

	return status, err
}
