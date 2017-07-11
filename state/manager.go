package state

import (
	"errors"
	"fmt"
	"os"
	"github.com/go-kit/kit/log"
)

type StateManager interface {
	Initialize(logger log.Logger, resource string) error
	Cleanup() error
	ListDefinitions(
		limit int, offset int, sortBy string,
		order string, filters map[string]string,
		envFilters map[string]string) (DefinitionList, error)
	GetDefinition(definitionID string) (Definition, error)
	UpdateDefinition(definitionID string, updates Definition) error
	CreateDefinition(d Definition) error
	DeleteDefinition(definitionID string) error

	ListRuns(limit int, offset int, sortBy string,
		order string, filters map[string]string,
		envFilters map[string]string) (RunList, error)

	GetRun(runID string) (Run, error)
	CreateRun(r Run) error
	UpdateRun(runID string, updates Run) error
}

func NewStateManager(logger log.Logger, name string) (StateManager, error) {
	switch name {
	case "postgres":
		pgm := &SQLStateManager{}
		return pgm, pgm.Initialize(logger, os.Getenv("DATABASE_URL"))
	default:
		return nil, errors.New(fmt.Sprintf("No StateManager named [%s] was found", name))
	}
}
