package services

import (
	"fmt"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/execution/engine"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/xeipuuv/gojsonschema"
	"strings"
)

//
// DefinitionService defines an interface for operations involving
// definitions
// * Like the ExecutionService, is an intermediary layer between state and the execution engine
//
type DefinitionService interface {
	Create(definition *state.Definition) (state.Definition, error)
	Get(definitionID string) (state.Definition, error)
	GetByAlias(alias string) (state.Definition, error)
	List(limit int, offset int, sortBy string,
		order string, filters map[string][]string,
		envFilters map[string]string) (state.DefinitionList, error)
	Update(definitionID string, updates state.Definition) (state.Definition, error)
	Delete(definitionID string) error

	// Metadata oriented
	ListGroups(limit int, offset int, name *string) (state.GroupsList, error)
	ListTags(limit int, offset int, name *string) (state.TagsList, error)
}

type definitionService struct {
	sm                 state.Manager
	ecsExecutionEngine engine.Engine
}

//
// NewDefinitionService configures and returns a DefinitionService
//
func NewDefinitionService(conf config.Config, ecsExecutionEngine engine.Engine, stateManager state.Manager) (DefinitionService, error) {
	ds := definitionService{sm: stateManager, ecsExecutionEngine: ecsExecutionEngine}
	return &ds, nil
}

//
// Create fully initialize and save the new definition
// * Allocates new definition id
// * Defines definition with execution engine
// * Stores definition using state manager
//
func (ds *definitionService) Create(definition *state.Definition) (state.Definition, error) {
	// Basic validation.
	if valid, reasons := definition.IsValid(); !valid {
		return state.Definition{}, exceptions.MalformedInput{ErrorString: strings.Join(reasons, "\n")}
	}

	// Validate unique alias.
	exists, err := ds.aliasExists(definition.Alias)
	if err != nil {
		return state.Definition{}, err
	}

	if exists {
		return state.Definition{}, exceptions.ConflictingResource{
			ErrorString: fmt.Sprintf("definition with alias [%s] aleady exists", definition.Alias)}
	}

	// If the task was created with a template_id, ensure that 1. the template
	// exists and 2. the `template_payload` conforms to the template's JSON
	// schema.
	if len(definition.TemplateID) > 0 {
		// Retrieve template.
		dt, err := ds.sm.GetDefinitionTemplateByID(definition.TemplateID)

		if err != nil {
			return state.Definition{}, err
		}

		schemaLoader := gojsonschema.NewStringLoader(dt.Schema)
		documentLoader := gojsonschema.NewStringLoader(fmt.Sprintf("%v", definition.TemplatePayload))

		// Perform validation.
		if _, err = gojsonschema.Validate(schemaLoader, documentLoader); err != nil {
			return state.Definition{}, err
		}
	}

	// Attach definition id here
	definitionID, err := state.NewDefinitionID(*definition)
	if err != nil {
		return state.Definition{}, err
	}
	definition.DefinitionID = definitionID
	defined, err := ds.ecsExecutionEngine.Define(*definition)

	if err != nil {
		return state.Definition{}, err
	}
	return defined, ds.sm.CreateDefinition(defined)
}

func (ds *definitionService) aliasExists(alias string) (bool, error) {
	// Short circuit, to check if alias already exists
	dl, err := ds.sm.ListDefinitions(
		1024, 0, "alias", "asc", map[string][]string{"alias": {alias}}, nil)

	if err != nil {
		return false, err
	}

	for _, def := range dl.Definitions {
		if def.Alias == alias {
			return true, nil
		}
	}
	return false, nil
}

//
// Get returns the definition specified by definitionID
//
func (ds *definitionService) Get(definitionID string) (state.Definition, error) {
	return ds.sm.GetDefinition(definitionID)
}

func (ds *definitionService) GetByAlias(alias string) (state.Definition, error) {
	return ds.sm.GetDefinitionByAlias(alias)
}

// List lists definitions
func (ds *definitionService) List(limit int, offset int, sortBy string,
	order string, filters map[string][]string,
	envFilters map[string]string) (state.DefinitionList, error) {
	return ds.sm.ListDefinitions(limit, offset, sortBy, order, filters, envFilters)
}

// UpdateStatus updates the definition specified by definitionID with the given updates
func (ds *definitionService) Update(definitionID string, updates state.Definition) (state.Definition, error) {
	definition, err := ds.sm.GetDefinition(definitionID)
	if err != nil {
		return definition, err
	}

	definition.UpdateWith(updates)
	defined, err := ds.ecsExecutionEngine.Define(definition)
	if err != nil {
		return definition, err
	}

	if definition.AdaptiveResourceAllocation != nil {
		defined.AdaptiveResourceAllocation = definition.AdaptiveResourceAllocation
	}

	return ds.sm.UpdateDefinition(definitionID, defined)
}

// Delete deletes and deregisters the definition specified by definitionID
func (ds *definitionService) Delete(definitionID string) error {
	definition, err := ds.sm.GetDefinition(definitionID)
	if err != nil {
		return err
	}
	if err = ds.ecsExecutionEngine.Deregister(definition); err != nil {
		return err
	}
	return ds.sm.DeleteDefinition(definitionID)
}

func (ds *definitionService) ListGroups(limit int, offset int, name *string) (state.GroupsList, error) {
	return ds.sm.ListGroups(limit, offset, name)
}

func (ds *definitionService) ListTags(limit int, offset int, name *string) (state.TagsList, error) {
	return ds.sm.ListTags(limit, offset, name)
}
