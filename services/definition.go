package services

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/execution/engine"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

type DefinitionService interface {
	Create(definition *state.Definition) (state.Definition, error)
	Get(definitionID string) (state.Definition, error)
	List(limit int, offset int, sortBy string,
		order string, filters map[string]string,
		envFilters map[string]string) (state.DefinitionList, error)
	Update(definitionID string, updates state.Definition) error
	Delete(definitionID string) error
}

type definitionService struct {
	sm state.Manager
	ee engine.Engine
}

func NewDefinitionService(conf config.Config, ee engine.Engine, sm state.Manager) (DefinitionService, error) {
	ds := definitionService{sm: sm, ee: ee}
	return &ds, nil
}

func (ds *definitionService) Create(definition *state.Definition) (state.Definition, error) {
	if valid, reasons := definition.IsValid(); !valid {
		return state.Definition{}, exceptions.MalformedInput{strings.Join(reasons, "\n")}
	}

	// Attach definition id here
	definitionID, err := state.NewDefinitionID(*definition)
	if err != nil {
		return state.Definition{}, err
	}
	definition.DefinitionID = definitionID

	defined, err := ds.ee.Define(*definition)
	if err != nil {
		return state.Definition{}, err
	}
	return defined, ds.sm.CreateDefinition(defined)
}

func (ds *definitionService) Get(definitionID string) (state.Definition, error) {
	return state.Definition{}, nil
}

func (ds *definitionService) List(limit int, offset int, sortBy string,
	order string, filters map[string]string,
	envFilters map[string]string) (state.DefinitionList, error) {
	return state.DefinitionList{}, nil
}

func (ds *definitionService) Update(definitionID string, updates state.Definition) error {
	return nil
}

func (ds *definitionService) Delete(definitionID string) error {
	return nil
}
