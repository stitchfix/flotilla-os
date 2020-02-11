package services

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

// TemplateService defines an interface for operations involving templates.
type TemplateService interface {
	Get(id string) (state.Template, error)
	List(limit int, offset int, sortBy string, order string) (state.TemplateList, error)
	Create(tpl *state.Template) (state.Template, error)
}

type templateService struct {
	sm state.Manager
}

// NewTemplateService configures and returns a TemplateService
func NewTemplateService(conf config.Config, sm state.Manager) (TemplateService, error) {
	ts := templateService{sm: sm}
	return &ts, nil
}

// Create fully initialize and save the new template.
func (ts *templateService) Create(t *state.Template) (state.Template, error) {
	if valid, reasons := t.IsValid(); !valid {
		return state.Template{}, exceptions.MalformedInput{ErrorString: strings.Join(reasons, "\n")}
	}

	// Attach template id.
	templateID, err := state.NewTemplateID(*t)
	if err != nil {
		return state.Template{}, err
	}
	t.TemplateID = templateID

	return *t, ts.sm.CreateTemplate(*t)
}

// Get returns the template specified by id.
func (ts *templateService) Get(id string) (state.Template, error) {
	return ts.sm.GetTemplate(id)
}

// List lists templates.
func (ts *templateService) List(limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	return ts.sm.ListTemplates(limit, offset, sortBy, order)
}
