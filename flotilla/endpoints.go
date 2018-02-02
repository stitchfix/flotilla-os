package flotilla

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
)

type endpoints struct {
	executionService  services.ExecutionService
	definitionService services.DefinitionService
	logService        services.LogService
}

type listRequest struct {
	limit      int
	offset     int
	sortBy     string
	order      string
	filters    map[string][]string
	envFilters map[string]string
}

type launchRequest struct {
	ClusterName string         `json:"cluster"`
	Env         *state.EnvList `json:"env"`
}

type launchRequestV2 struct {
	RunTags RunTags `json:"run_tags"`
	*launchRequest
}

//
// RunTags represents which user is responsible for a task run
//
type RunTags struct {
	OwnerEmail string `json:"owner_email"`
	TeamName   string `json:"team_name"`
	OwnerID    string `json:"owner_id"`
}

func (ep *endpoints) getURLParam(v url.Values, key string, defaultValue string) string {
	val, ok := v[key]
	if ok && len(val) > 0 {
		return val[0]
	}
	return defaultValue
}

func (ep *endpoints) getFilters(params url.Values, nonFilters map[string]bool) (map[string][]string, map[string]string) {
	filters := make(map[string][]string)
	envFilters := make(map[string]string)
	for k, v := range params {
		if !nonFilters[k] && len(v) > 0 {
			// Env filters have the "env" key and are "|" separated key-value pairs
			//
			// eg. env=FOO|BAR&env=CUPCAKE|SPRINKLES
			//
			if k == "env" {
				for _, kv := range v {
					split := strings.Split(kv, "|")
					if len(split) == 2 {
						envFilters[split[0]] = split[1]
					}
				}
			} else {
				filters[k] = v
			}
		}
	}
	return filters, envFilters
}

func (ep *endpoints) decodeListRequest(r *http.Request) listRequest {
	var lr listRequest
	params := r.URL.Query()

	lr.limit, _ = strconv.Atoi(ep.getURLParam(params, "limit", "1024"))
	lr.offset, _ = strconv.Atoi(ep.getURLParam(params, "offset", "0"))
	lr.sortBy = ep.getURLParam(params, "sort_by", "group_name")
	lr.order = ep.getURLParam(params, "order", "asc")
	lr.filters, lr.envFilters = ep.getFilters(params, map[string]bool{
		"limit":   true,
		"offset":  true,
		"sort_by": true,
		"order":   true,
	})
	return lr
}

func (ep *endpoints) decodeRequest(r *http.Request, entity interface{}) error {
	return json.NewDecoder(r.Body).Decode(entity)
}

func (ep endpoints) encodeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err.(type) {
	case exceptions.MalformedInput:
		w.WriteHeader(http.StatusBadRequest)
	case exceptions.ConflictingResource:
		w.WriteHeader(http.StatusConflict)
	case exceptions.MissingResource:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func (ep *endpoints) encodeResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

func (ep *endpoints) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	definitionList, err := ep.definitionService.List(
		lr.limit, lr.offset, lr.sortBy, lr.order, lr.filters, lr.envFilters)
	if definitionList.Definitions == nil {
		definitionList.Definitions = []state.Definition{}
	}
	if err != nil {
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["total"] = definitionList.Total
		response["definitions"] = definitionList.Definitions
		response["limit"] = lr.limit
		response["offset"] = lr.offset
		response["sort_by"] = lr.sortBy
		response["order"] = lr.order
		response["env_filters"] = lr.envFilters
		for k, v := range lr.filters {
			response[k] = v
		}
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) GetDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	definition, err := ep.definitionService.Get(vars["definition_id"])
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, definition)
	}
}

func (ep *endpoints) GetDefinitionByAlias(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	definition, err := ep.definitionService.GetByAlias(vars["alias"])
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, definition)
	}
}

func (ep *endpoints) CreateDefinition(w http.ResponseWriter, r *http.Request) {
	var definition state.Definition
	err := ep.decodeRequest(r, &definition)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	created, err := ep.definitionService.Create(&definition)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, created)
	}
}

func (ep *endpoints) UpdateDefinition(w http.ResponseWriter, r *http.Request) {
	var definition state.Definition
	err := ep.decodeRequest(r, &definition)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	vars := mux.Vars(r)
	updated, err := ep.definitionService.Update(vars["definition_id"], definition)

	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, updated)
	}
}

func (ep *endpoints) DeleteDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := ep.definitionService.Delete(vars["definition_id"])
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, map[string]bool{"deleted": true})
	}
}

func (ep *endpoints) ListRuns(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	vars := mux.Vars(r)
	definitionID, ok := vars["definition_id"]
	if ok {
		lr.filters["definition_id"] = []string{definitionID}
	}

	runList, err := ep.executionService.List(
		lr.limit, lr.offset, lr.order, lr.sortBy, lr.filters, lr.envFilters)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["total"] = runList.Total
		response["history"] = runList.Runs
		response["limit"] = lr.limit
		response["offset"] = lr.offset
		response["sort_by"] = lr.sortBy
		response["order"] = lr.order
		response["env_filters"] = lr.envFilters
		for k, v := range lr.filters {
			response[k] = v
		}
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) GetRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := ep.executionService.Get(vars["run_id"])
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRun(w http.ResponseWriter, r *http.Request) {
	var lr launchRequest
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	vars := mux.Vars(r)
	run, err := ep.executionService.Create(vars["definition_id"], lr.ClusterName, lr.Env, "v1-unknown")
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRunV2(w http.ResponseWriter, r *http.Request) {
	var lr launchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	if len(lr.RunTags.OwnerEmail) == 0 || len(lr.RunTags.TeamName) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_email] and [team_name]")})
		return
	}

	vars := mux.Vars(r)
	run, err := ep.executionService.Create(vars["definition_id"], lr.ClusterName, lr.Env, lr.RunTags.OwnerEmail)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRunV4(w http.ResponseWriter, r *http.Request) {
	var lr launchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	if len(lr.RunTags.OwnerID) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_id]")})
		return
	}

	vars := mux.Vars(r)
	run, err := ep.executionService.Create(vars["definition_id"], lr.ClusterName, lr.Env, lr.RunTags.OwnerID)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRunByAlias(w http.ResponseWriter, r *http.Request) {
	var lr launchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	if len(lr.RunTags.OwnerID) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_id]")})
		return
	}

	vars := mux.Vars(r)
	run, err := ep.executionService.CreateByAlias(vars["alias"], lr.ClusterName, lr.Env, lr.RunTags.OwnerID)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) StopRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := ep.executionService.Terminate(vars["run_id"])
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, map[string]bool{"terminated": true})
	}
}

func (ep *endpoints) UpdateRun(w http.ResponseWriter, r *http.Request) {
	var run state.Run
	err := ep.decodeRequest(r, &run)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	vars := mux.Vars(r)
	err = ep.executionService.UpdateStatus(vars["run_id"], run.Status, run.ExitCode)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, map[string]bool{"updated": true})
	}
}

func (ep *endpoints) GetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := r.URL.Query()

	lastSeen := ep.getURLParam(params, "last_seen", "")
	logs, newLastSeen, err := ep.logService.Logs(vars["run_id"], &lastSeen)
	if err != nil {
		ep.encodeError(w, err)
		return
	}

	res := map[string]string{
		"log": logs,
	}
	if newLastSeen != nil {
		res["last_seen"] = *newLastSeen
	}
	ep.encodeResponse(w, res)
}

func (ep *endpoints) GetGroups(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	var name string
	if len(lr.filters["name"]) > 0 {
		name = lr.filters["name"][0]
	}

	groups, err := ep.definitionService.ListGroups(lr.limit, lr.offset, &name)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["total"] = groups.Total
		response["groups"] = groups.Groups
		response["limit"] = lr.limit
		response["offset"] = lr.offset
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) GetTags(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	var name string
	if len(lr.filters["name"]) > 0 {
		name = lr.filters["name"][0]
	}

	tags, err := ep.definitionService.ListTags(lr.limit, lr.offset, &name)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["total"] = tags.Total
		response["tags"] = tags.Tags
		response["limit"] = lr.limit
		response["offset"] = lr.offset
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) ListClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := ep.executionService.ListClusters()
	if err != nil {
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["clusters"] = clusters
		ep.encodeResponse(w, response)
	}
}
