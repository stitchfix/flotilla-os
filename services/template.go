package services

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

//
// TemplateService defines an interface for operations involving
// definitions
// * Like the ExecutionService, is an intermediary layer between state and the execution engine
//
type TemplateService interface {
	Create(definition *state.Template) (state.Template, error)
	Get(id string) (state.Template, error)
	List(limit int, offset int, sortBy string, order string) (state.TemplateList, error)
	Update(id string, updates state.Template) (state.Template, error)
	Delete(id string) error
}

type templateService struct {
	sm state.Manager
}

//
// NewTemplateService configures and returns a TemplateService
//
func NewTemplateService(conf config.Config, sm state.Manager) (TemplateService, error) {
	ts := templateService{sm: sm}
	return &ts, nil
}

//
// Create fully initialize and save the new template.
// * Allocates new definition id
// * Stores definition using state manager
//
func (ds *templateService) Create(t *state.Template) (state.Template, error) {
	if valid, reasons := t.IsValid(); !valid {
		return state.Template{}, exceptions.MalformedInput{ErrorString: strings.Join(reasons, "\n")}
	}

	// Attach definition id here
	templateID, err := state.NewTemplateID(*t)
	if err != nil {
		return state.Template{}, err
	}
	t.TemplateID = templateID

	return *t, ds.sm.CreateTemplate(*t)
}

//
// Get returns the definition specified by id
//
func (ds *templateService) Get(id string) (state.Template, error) {
	return ds.sm.GetTemplate(id)
}

// List lists definitions
func (ds *templateService) List(limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	return ds.sm.ListTemplates(limit, offset, sortBy, order)
}

// UpdateStatus updates the definition specified by id with the given updates
func (ds *templateService) Update(id string, updates state.Template) (state.Template, error) {
	tpl, err := ds.sm.GetTemplate(id)
	if err != nil {
		return tpl, err
	}

	// TODO: implement
	// tpl.UpdateWith(updates)

	return ds.sm.UpdateTemplate(id, tpl)
}

// Delete deletes the template specified by id
func (ds *templateService) Delete(id string) error {
	_, err := ds.sm.GetTemplate(id)
	if err != nil {
		return err
	}
	return ds.sm.DeleteTemplate(id)
}
