package services

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/state"
)

// TemplateService defines an interface for operations involving templates.
type TemplateService interface {
	GetByID(id string) (state.Template, error)
	GetLatestByName(templateName string) (bool, state.Template, error)
	List(limit int, offset int, sortBy string, order string) (state.TemplateList, error)
	ListLatestOnly(limit int, offset int, sortBy string, order string) (state.TemplateList, error)
	Create(tpl *state.CreateTemplateRequest) (state.CreateTemplateResponse, error)
}

type templateService struct {
	sm state.Manager
}

// NewTemplateService configures and returns a TemplateService.
func NewTemplateService(conf config.Config, sm state.Manager) (TemplateService, error) {
	ts := templateService{sm: sm}
	return &ts, nil
}

// Create fully initialize and save the new template.
func (ts *templateService) Create(req *state.CreateTemplateRequest) (state.CreateTemplateResponse, error) {
	res := state.CreateTemplateResponse{
		DidCreate: false,
		Template:  state.Template{},
	}
	curr, err := ts.constructTemplateFromCreateTemplateRequest(req)

	// 1. Check validity.
	if valid, reasons := curr.IsValid(); !valid {
		return res, exceptions.MalformedInput{ErrorString: strings.Join(reasons, "\n")}
	}

	// 2. Attach template id.
	templateID, err := state.NewTemplateID(curr)
	if err != nil {
		return res, err
	}
	curr.TemplateID = templateID

	// 3. Check if template name exists - if it does NOT, we will insert it into
	// the DB with a version number of 1. If it does, and if there are any
	// changed fields, then we will create a new row in the DB w/ the version
	// incremented by 1. If there are NO changed fields, then just return the
	// latest version.
	doesExist, prev, err := ts.sm.GetLatestTemplateByTemplateName(curr.TemplateName)

	if err != nil {
		return res, err
	}

	// No previous template with the same name; write it.
	if doesExist == false {
		curr.Version = 1
		res.Template = curr
		res.DidCreate = true
		return res, ts.sm.CreateTemplate(curr)
	}

	// Check if prev and curr are diff, if they are, write curr to DB (increment)
	// version number by 1. Otherwise, return prev.
	if ts.diff(prev, curr) == true {
		curr.Version = prev.Version + 1
		res.Template = curr
		res.DidCreate = true
		return res, ts.sm.CreateTemplate(curr)
	}

	res.Template = prev
	return res, nil
}

// Get returns the template specified by id.
func (ts *templateService) GetByID(id string) (state.Template, error) {
	return ts.sm.GetTemplateByID(id)
}

// Get returns the template specified by id.
func (ts *templateService) GetLatestByName(templateName string) (bool, state.Template, error) {
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

// diff performs a diff between all fields (except for TemplateName and
// Version) of two templates.
func (ts *templateService) diff(prev state.Template, curr state.Template) bool {
	if prev.TemplateName != curr.TemplateName {
		return true
	}
	if prev.CommandTemplate != curr.CommandTemplate {
		return true
	}
	if prev.Image != curr.Image {
		return true
	}
	if *prev.Memory != *curr.Memory {
		return true
	}
	if *prev.Gpu != *curr.Gpu {
		return true
	}
	if *prev.Cpu != *curr.Cpu {
		return true
	}

	if prev.Env != nil && curr.Env != nil {
		prevEnv := *prev.Env
		currEnv := *curr.Env
		if len(prevEnv) != len(currEnv) {
			return true
		}

		for i, e := range prevEnv {
			if e != currEnv[i] {
				return true
			}
		}
	}

	if *prev.Privileged != *curr.Privileged {
		return true
	}
	if *prev.AdaptiveResourceAllocation != *curr.AdaptiveResourceAllocation {
		return true
	}
	if prev.ContainerName != curr.ContainerName {
		return true
	}

	if prev.Ports != nil && curr.Ports != nil {
		prevPorts := *prev.Ports
		currPorts := *curr.Ports
		if len(prevPorts) != len(currPorts) {
			return true
		}

		for i, e := range prevPorts {
			if e != currPorts[i] {
				return true
			}
		}
	}

	if prev.Tags != nil && curr.Tags != nil {
		prevTags := *prev.Tags
		currTags := *curr.Tags
		if len(prevTags) != len(currTags) {
			return true
		}

		for i, e := range prevTags {
			if e != currTags[i] {
				return true
			}
		}
	}

	if reflect.DeepEqual(prev.Schema, curr.Schema) == false {
		fmt.Println("M")
		return true
	}

	return false
}

// constructTemplateFromCreateTemplateRequest takes a CreateTemplateRequest and
// dumps the requisite fields into a Template.
func (ts *templateService) constructTemplateFromCreateTemplateRequest(req *state.CreateTemplateRequest) (state.Template, error) {
	tpl := state.Template{}

	if len(req.TemplateName) > 0 {
		tpl.TemplateName = req.TemplateName
	}
	if req.Schema != nil {
		tpl.Schema = req.Schema
	}
	if len(req.CommandTemplate) > 0 {
		tpl.CommandTemplate = req.CommandTemplate
	}
	if len(req.Image) > 0 {
		tpl.Image = req.Image
	}
	if req.Memory != nil {
		tpl.Memory = req.Memory
	}
	if req.Gpu != nil {
		tpl.Gpu = req.Gpu
	}
	if req.Cpu != nil {
		tpl.Cpu = req.Cpu
	}
	if req.Env != nil {
		tpl.Env = req.Env
	}
	if req.Privileged != nil {
		tpl.Privileged = req.Privileged
	}
	if req.AdaptiveResourceAllocation != nil {
		tpl.AdaptiveResourceAllocation = req.AdaptiveResourceAllocation
	}
	if len(req.ContainerName) > 0 {
		tpl.ContainerName = req.ContainerName
	}
	if req.Ports != nil {
		tpl.Ports = req.Ports
	}
	if req.Tags != nil {
		tpl.Tags = req.Tags
	}

	return tpl, nil
}
