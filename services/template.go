package services

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
	"strings"
)

// TemplateService defines an interface for operations involving templates.
type TemplateService interface {
	GetByID(id string) (state.Template, error)
	GetLatestByName(templateName string) (state.Template, error)
	List(limit int, offset int, sortBy string, order string) (state.TemplateList, error)
	ListLatestOnly(limit int, offset int, sortBy string, order string) (state.TemplateList, error)
	Create(tpl *state.CreateTemplateRequest) (state.Template, error)
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
func (ts *templateService) Create(t *state.CreateTemplateRequest) (state.Template, error) {
	curr := state.Template{
		TemplateName:    t.TemplateName,
		Schema:          t.Schema,
		CommandTemplate: t.CommandTemplate,
		ExecutableResources: state.ExecutableResources{
			Image:                      t.Image,
			Memory:                     t.Memory,
			Gpu:                        t.Gpu,
			Cpu:                        t.Cpu,
			Env:                        t.Env,
			Privileged:                 t.Privileged,
			AdaptiveResourceAllocation: t.AdaptiveResourceAllocation,
			ContainerName:              t.ContainerName,
			Ports:                      t.Ports,
			Tags:                       t.Tags,
		},
	}

	// 1. Check validity.
	if valid, reasons := curr.IsValid(); !valid {
		return curr, exceptions.MalformedInput{ErrorString: strings.Join(reasons, "\n")}
	}

	// 2. Attach template id.
	templateID, err := state.NewTemplateID(curr)
	if err != nil {
		return state.Template{}, err
	}
	curr.TemplateID = templateID

	// 3. Check if template name exists - if it does NOT, we will insert it into
	// the DB with a version number of 1. If it does, and if there are any
	// changed fields, then we will create a new row in the DB w/ the version
	// incremented by 1. If there are NO changed fields, then just return the
	// latest version.
	prev, err := ts.sm.GetLatestTemplateByTemplateName(t.TemplateName)

	// No previous template with the same name; write it.
	if &prev == nil {
		curr.Version = 1
		return curr, ts.sm.CreateTemplate(curr)
	}

	// Has changes.
	if ts.diff(curr, prev) == true {
		curr.Version = prev.Version + 1
		return curr, ts.sm.CreateTemplate(curr)
	}

	return prev, nil
}

// Get returns the template specified by id.
func (ts *templateService) GetByID(id string) (state.Template, error) {
	return ts.sm.GetTemplateByID(id)
}

// Get returns the template specified by id.
func (ts *templateService) GetLatestByName(templateName string) (state.Template, error) {
	return ts.sm.GetLatestTemplateByTemplateName(templateName)
}

// List lists templates.
func (ts *templateService) List(limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	return ts.sm.ListTemplates(limit, offset, sortBy, order)
}

// List lists templates.
func (ts *templateService) ListLatestOnly(limit int, offset int, sortBy string, order string) (state.TemplateList, error) {
	return ts.sm.ListTemplatesLatestOnly(limit, offset, sortBy, order)
}

// List lists templates.
func (ts *templateService) diff(curr state.Template, prev state.Template) bool {
	return false
}
