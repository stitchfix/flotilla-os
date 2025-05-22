package state

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/log"
)

// Manager interface for CRUD operations
// on definitions and runs
type Manager interface {
	Name() string
	Initialize(conf config.Config) error
	Cleanup() error
	ListDefinitions(
		ctx context.Context,
		limit int, offset int, sortBy string,
		order string, filters map[string][]string,
		envFilters map[string]string) (DefinitionList, error)
	GetDefinition(ctx context.Context, definitionID string) (Definition, error)
	GetDefinitionByAlias(ctx context.Context, alias string) (Definition, error)
	UpdateDefinition(ctx context.Context, definitionID string, updates Definition) (Definition, error)
	CreateDefinition(ctx context.Context, d Definition) error
	DeleteDefinition(ctx context.Context, definitionID string) error

	ListRuns(ctx context.Context, limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string, engines []string) (RunList, error)
	EstimateRunResources(ctx context.Context, executableID string, commandHash string) (TaskResources, error)
	EstimateExecutorCount(ctx context.Context, executableID string, commandHash string) (int64, error)
	ExecutorOOM(ctx context.Context, executableID string, commandHash string) (bool, error)
	DriverOOM(ctx context.Context, executableID string, commandHash string) (bool, error)

	GetRun(ctx context.Context, runID string) (Run, error)
	CreateRun(ctx context.Context, r Run) error
	UpdateRun(ctx context.Context, runID string, updates Run) (Run, error)

	ListGroups(ctx context.Context, limit int, offset int, name *string) (GroupsList, error)
	ListTags(ctx context.Context, limit int, offset int, name *string) (TagsList, error)

	ListWorkers(ctx context.Context, engine string) (WorkersList, error)
	BatchUpdateWorkers(ctx context.Context, updates []Worker) (WorkersList, error)
	GetWorker(ctx context.Context, workerType string, engine string) (Worker, error)
	UpdateWorker(ctx context.Context, workerType string, updates Worker) (Worker, error)

	GetExecutableByTypeAndID(ctx context.Context, executableType ExecutableType, executableID string) (Executable, error)

	GetTemplateByID(ctx context.Context, templateID string) (Template, error)
	GetLatestTemplateByTemplateName(ctx context.Context, templateName string) (bool, Template, error)
	GetTemplateByVersion(ctx context.Context, templateName string, templateVersion int64) (bool, Template, error)
	ListTemplates(ctx context.Context, limit int, offset int, sortBy string, order string) (TemplateList, error)
	ListTemplatesLatestOnly(ctx context.Context, limit int, offset int, sortBy string, order string) (TemplateList, error)
	CreateTemplate(ctx context.Context, t Template) error

	ListFailingNodes(ctx context.Context) (NodeList, error)
	GetPodReAttemptRate(ctx context.Context) (float32, error)
	GetNodeLifecycle(ctx context.Context, executableID string, commandHash string) (string, error)
	GetTaskHistoricalRuntime(ctx context.Context, executableID string, runId string) (float32, error)
	CheckIdempotenceKey(ctx context.Context, idempotenceKey string) (string, error)

	GetRunByEMRJobId(ctx context.Context, emrJobId string) (Run, error)
	GetResources(ctx context.Context, runID string) (Run, error)
	ListClusterStates(ctx context.Context) ([]ClusterMetadata, error)
	UpdateClusterMetadata(ctx context.Context, cluster ClusterMetadata) error
	DeleteClusterMetadata(ctx context.Context, clusterID string) error
	GetClusterByID(ctx context.Context, clusterID string) (ClusterMetadata, error)
	GetRunStatus(ctx context.Context, runID string) (RunStatus, error)
}

// NewStateManager sets up and configures a new statemanager
// - if no `state_manager` is configured, will use postgres
func NewStateManager(conf config.Config, logger log.Logger) (Manager, error) {
	name := "postgres"
	if conf.IsSet("state_manager") {
		name = conf.GetString("state_manager")
	}

	switch name {
	case "postgres":
		pgm := &SQLStateManager{log: logger}
		err := pgm.Initialize(conf)
		if err != nil {
			return nil, errors.Wrap(err, "problem initializing SQLStateManager")
		}
		return pgm, nil
	default:
		return nil, errors.Errorf("state.Manager named [%s] not found", name)
	}
}
