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

	ListRuns(limit int, offset int, sortBy string,
		order string, filters map[string][]string,
		envFilters map[string]string) (RunList, error)

	GetRun(runID string) (Run, error)
	CreateRun(r Run) error
	UpdateRun(runID string, updates Run) (Run, error)

	ListGroups(limit int, offset int, name *string) (GroupsList, error)
	ListTags(limit int, offset int, name *string) (TagsList, error)
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
