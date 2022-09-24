package state

import (
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
)

//
// Manager interface for CRUD operations on
// on definitions and runs
//
type Manager interface {
	Name() string
	Initialize(conf config.Config) error
	Cleanup() error
	ListDefinitions(
		limit int, offset int, sortBy string,
		order string, filters map[string][]string,
		envFilters map[string]string) (DefinitionList, error)
	GetDefinition(definitionID string) (Definition, error)
	GetDefinitionByAlias(alias string) (Definition, error)
	UpdateDefinition(definitionID string, updates Definition) (Definition, error)
	CreateDefinition(d Definition) error
	DeleteDefinition(definitionID string) error

	ListRuns(limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string, engines []string) (RunList, error)
	EstimateRunResources(executableID string, commandHash string) (TaskResources, error)
	EstimateExecutorCount(executableID string, commandHash string) (int64, error)
	ExecutorOOM(executableID string, commandHash string) (bool, error)
	DriverOOM(executableID string, commandHash string) (bool, error)

	GetRun(runID string) (Run, error)
	CreateRun(r Run) error
	UpdateRun(runID string, updates Run) (Run, error)

	ListGroups(limit int, offset int, name *string) (GroupsList, error)
	ListTags(limit int, offset int, name *string) (TagsList, error)

	ListWorkers(engine string) (WorkersList, error)
	BatchUpdateWorkers(updates []Worker) (WorkersList, error)
	GetWorker(workerType string, engine string) (Worker, error)
	UpdateWorker(workerType string, updates Worker) (Worker, error)

	GetExecutableByTypeAndID(executableType ExecutableType, executableID string) (Executable, error)

	GetTemplateByID(templateID string) (Template, error)
	GetLatestTemplateByTemplateName(templateName string) (bool, Template, error)
	GetTemplateByVersion(templateName string, templateVersion int64) (bool, Template, error)
	ListTemplates(limit int, offset int, sortBy string, order string) (TemplateList, error)
	ListTemplatesLatestOnly(limit int, offset int, sortBy string, order string) (TemplateList, error)
	CreateTemplate(t Template) error

	ListFailingNodes() (NodeList, error)
	GetPodReAttemptRate() (float32, error)
	GetNodeLifecycle(executableID string, commandHash string) (string, error)
	GetTaskHistoricalRuntime(executableID string, runId string) (float32, error)
	CheckIdempotenceKey(idempotenceKey string) (string, error)

	GetRunByEMRJobId(string) (Run, error)
}

//
// NewStateManager sets up and configures a new statemanager
// - if no `state_manager` is configured, will use postgres
//
func NewStateManager(conf config.Config) (Manager, error) {
	name := "postgres"
	if conf.IsSet("state_manager") {
		name = conf.GetString("state_manager")
	}

	switch name {
	case "postgres":
		pgm := &SQLStateManager{}
		err := pgm.Initialize(conf)
		if err != nil {
			return nil, errors.Wrap(err, "problem initializing SQLStateManager")
		}
		return pgm, nil
	default:
		return nil, errors.Errorf("state.Manager named [%s] not found", name)
	}
}
