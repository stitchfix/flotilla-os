package services

import (
	"context"
	"fmt"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

//
// DefinitionService defines an interface for operations involving
// definitions
// * Like the ExecutionService, is an intermediary layer between state and the execution engine
//
type DefinitionService interface {
	Create(ctx context.Context, definition *state.Definition) (state.Definition, error)
	Get(ctx context.Context, definitionID string) (state.Definition, error)
	GetByAlias(ctx context.Context, alias string) (state.Definition, error)
	List(ctx context.Context, limit int, offset int, sortBy string,
		order string, filters map[string][]string,
		envFilters map[string]string) (state.DefinitionList, error)
	Update(ctx context.Context, definitionID string, updates state.Definition) (state.Definition, error)
	Delete(ctx context.Context, definitionID string) error

	// Metadata oriented
	ListGroups(ctx context.Context, limit int, offset int, name *string) (state.GroupsList, error)
	ListTags(ctx context.Context, limit int, offset int, name *string) (state.TagsList, error)
}

type definitionService struct {
	sm state.Manager
}

//
// NewDefinitionService configures and returns a DefinitionService
//
func NewDefinitionService(stateManager state.Manager) (DefinitionService, error) {
	ds := definitionService{sm: stateManager}
	return &ds, nil
}

//
// Create fully initialize and save the new definition
// * Allocates new definition id
// * Defines definition with execution engine
// * Stores definition using state manager
//
func (ds *definitionService) Create(ctx context.Context, definition *state.Definition) (state.Definition, error) {
	if valid, reasons := definition.IsValid(); !valid {
		return state.Definition{}, exceptions.MalformedInput{strings.Join(reasons, "\n")}
	}

	exists, err := ds.aliasExists(ctx, definition.Alias)
	if err != nil {
		return state.Definition{}, err
	}

	if exists {
		return state.Definition{}, exceptions.ConflictingResource{
			fmt.Sprintf("definition with alias [%s] aleady exists", definition.Alias)}
	}
	// Attach definition id here
	definitionID, err := state.NewDefinitionID(*definition)
	if err != nil {
		return state.Definition{}, err
	}
	definition.DefinitionID = definitionID
	return *definition, ds.sm.CreateDefinition(ctx, *definition)
}

func (ds *definitionService) aliasExists(ctx context.Context, alias string) (bool, error) {
	// Short circuit, to check if alias already exists
	dl, err := ds.sm.ListDefinitions(
		ctx, 1024, 0, "alias", "asc", map[string][]string{"alias": {alias}}, nil)

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
func (ds *definitionService) Get(ctx context.Context, definitionID string) (state.Definition, error) {
	return ds.sm.GetDefinition(ctx, definitionID)
}

func (ds *definitionService) GetByAlias(ctx context.Context, alias string) (state.Definition, error) {
	return ds.sm.GetDefinitionByAlias(ctx, alias)
}

// List lists definitions
func (ds *definitionService) List(ctx context.Context, limit int, offset int, sortBy string,
	order string, filters map[string][]string,
	envFilters map[string]string) (state.DefinitionList, error) {
	return ds.sm.ListDefinitions(ctx, limit, offset, sortBy, order, filters, envFilters)
}

// UpdateStatus updates the definition specified by definitionID with the given updates
func (ds *definitionService) Update(ctx context.Context, definitionID string, updates state.Definition) (state.Definition, error) {
	definition, err := ds.sm.GetDefinition(ctx, definitionID)
	if err != nil {
		return definition, err
	}

	definition.UpdateWith(updates)
	return ds.sm.UpdateDefinition(ctx, definitionID, definition)
}

// Delete deletes and deregisters the definition specified by definitionID
func (ds *definitionService) Delete(ctx context.Context, definitionID string) error {
	return ds.sm.DeleteDefinition(ctx, definitionID)
}

func (ds *definitionService) ListGroups(ctx context.Context, limit int, offset int, name *string) (state.GroupsList, error) {
	return ds.sm.ListGroups(ctx, limit, offset, name)
}

func (ds *definitionService) ListTags(ctx context.Context, limit int, offset int, name *string) (state.TagsList, error) {
	return ds.sm.ListTags(ctx, limit, offset, name)
}
