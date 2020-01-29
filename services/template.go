package services

import (
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/state"
)

//
// TemplateService defines an interface for operations involving definition
// templates.
//
type TemplateService interface {
	List(limit int, offset int, latestOnly bool) (state.DefinitionTemplateList, error)
	GetByID(id string) (state.DefinitionTemplate, error)
}

type templateService struct {
	sm state.Manager
}

//
// NewDefinitionTemplateService configures and returns a DefinitionTemplateService
//
func NewDefinitionTemplateService(conf config.Config, sm state.Manager) (TemplateService, error) {
	ts := templateService{sm: sm}
	return &ts, nil
}

func (ts *templateService) List(limit int, offset int, latestOnly bool) (state.DefinitionTemplateList, error) {
	return ts.sm.ListDefinitionTemplates(limit, offset, latestOnly)
}

func (ts *templateService) GetByID(id string) (state.DefinitionTemplate, error) {
	return ts.sm.GetDefinitionTemplateByID(id)
}
