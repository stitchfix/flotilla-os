package flotilla

import (
	"github.com/stitchfix/flotilla-os/services"
	"net/http"
)

type endpoints struct {
	executionService  services.ExecutionService
	definitionService services.DefinitionService
}

func (ep *endpoints) ListDefinitions(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) GetDefinition(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) CreateDefinition(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) UpdateDefinition(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) DeleteDefinition(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) ListRuns(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) GetRun(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) CreateRun(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) StopRun(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) GetLogs(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) GetGroups(w http.ResponseWriter, r *http.Request) {

}
