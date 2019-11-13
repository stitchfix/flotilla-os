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
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
)

type endpoints struct {
	executionService  services.ExecutionService
	definitionService services.DefinitionService
	ecsLogService     services.LogService
	eksLogService     services.LogService
	workerService     services.WorkerService
	logger            flotillaLog.Logger
}

type listRequest struct {
	limit      int
	offset     int
	sortBy     string
	order      string
	filters    map[string][]string
	envFilters map[string]string
}

type LaunchRequest struct {
	ClusterName string         `json:"cluster"`
	Env         *state.EnvList `json:"env"`
}

type LaunchRequestV2 struct {
	RunTags          RunTags `json:"run_tags"`
	Command          *string
	Memory           *int64
	Cpu              *int64
	Engine           *string
	NodeLifecycle    *string `json:"node_lifecycle"`
	EphemeralStorage *int64 `json:"ephemeral_storage"`
	*LaunchRequest
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
		ep.logger.Log(
			"message", "problem listing definitions",
			"operation", "ListDefinitions",
			"error", fmt.Sprintf("%+v", err))
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
		ep.logger.Log(
			"message", "problem getting definitions",
			"operation", "GetDefinition",
			"error", fmt.Sprintf("%+v", err),
			"definition_id", vars["definition_id"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, definition)
	}
}

func (ep *endpoints) GetDefinitionByAlias(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	definition, err := ep.definitionService.GetByAlias(vars["alias"])
	if err != nil {
		ep.logger.Log(
			"message", "problem getting definition by alias",
			"operation", "GetDefinitionByAlias",
			"error", fmt.Sprintf("%+v", err),
			"alias", vars["alias"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, definition)
	}
}

func (ep *endpoints) CreateDefinition(w http.ResponseWriter, r *http.Request) {
	var definition state.Definition
	err := ep.decodeRequest(r, &definition)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	created, err := ep.definitionService.Create(&definition)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating definition",
			"operation", "CreateDefinition",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, created)
	}
}

func (ep *endpoints) UpdateDefinition(w http.ResponseWriter, r *http.Request) {
	var definition state.Definition
	err := ep.decodeRequest(r, &definition)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	vars := mux.Vars(r)
	updated, err := ep.definitionService.Update(vars["definition_id"], definition)

	if err != nil {
		ep.logger.Log(
			"message", "problem updating definition",
			"operation", "UpdateDefinition",
			"error", fmt.Sprintf("%+v", err),
			"definition_id", vars["definition_id"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, updated)
	}
}

func (ep *endpoints) DeleteDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := ep.definitionService.Delete(vars["definition_id"])
	if err != nil {
		ep.logger.Log(
			"message", "problem deleting definition",
			"operation", "DeleteDefinition",
			"error", fmt.Sprintf("%+v", err),
			"definition_id", vars["definition_id"])
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
		ep.logger.Log(
			"message", "problem listing runs",
			"operation", "ListRuns",
			"error", fmt.Sprintf("%+v", err))
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
		ep.logger.Log(
			"message", "problem getting run",
			"operation", "GetRun",
			"error", fmt.Sprintf("%+v", err),
			"run_id", vars["run_id"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRun(w http.ResponseWriter, r *http.Request) {
	var lr LaunchRequest
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	vars := mux.Vars(r)
	engine := state.DefaultEngine
	run, err := ep.executionService.Create(vars["definition_id"], lr.ClusterName, lr.Env, "v1-unknown", nil, nil, nil, &engine, nil, nil)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating run",
			"operation", "CreateRun",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRunV2(w http.ResponseWriter, r *http.Request) {
	var lr LaunchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	if len(lr.RunTags.OwnerEmail) == 0 || len(lr.RunTags.TeamName) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_email] and [team_name]")})
		return
	}

	vars := mux.Vars(r)
	if lr.Engine != nil {
		if !stringInSlice(*lr.Engine, state.Engines) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("engine must be [ecs, eks]")})
			return
		}
	} else {
		lr.Engine = &state.DefaultEngine
	}
	run, err := ep.executionService.Create(vars["definition_id"], lr.ClusterName, lr.Env, lr.RunTags.OwnerEmail, nil, nil, nil, lr.Engine, nil, nil)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating V2 run",
			"operation", "CreateRunV2",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func (ep *endpoints) CreateRunV4(w http.ResponseWriter, r *http.Request) {
	var lr LaunchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	if len(lr.RunTags.OwnerID) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_id]")})
		return
	}

	if lr.Engine != nil {
		if !stringInSlice(*lr.Engine, state.Engines) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("engine must be [ecs, eks] %s was specified", *lr.Engine)})
			return
		}
	} else {
		lr.Engine = &state.DefaultEngine
	}

	if lr.NodeLifecycle != nil {
		if !stringInSlice(*lr.NodeLifecycle, state.NodeLifeCycles) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("Nodelifecyle must be [normal, spot]")})
			return
		}
	} else {
		lr.NodeLifecycle = &state.DefaultLifecycle
	}
	vars := mux.Vars(r)
	run, err := ep.executionService.Create(vars["definition_id"], lr.ClusterName, lr.Env, lr.RunTags.OwnerID, lr.Command, lr.Memory, lr.Cpu, lr.Engine, lr.EphemeralStorage, lr.NodeLifecycle)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating V4 run",
			"operation", "CreateRunV4",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) CreateRunByAlias(w http.ResponseWriter, r *http.Request) {
	var lr LaunchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	if len(lr.RunTags.OwnerID) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_id]")})
		return
	}

	if lr.Engine != nil {
		if !stringInSlice(*lr.Engine, state.Engines) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("engine must be [ecs, eks]")})
			return
		}
	} else {
		lr.Engine = &state.DefaultEngine
	}

	if lr.NodeLifecycle != nil {
		if !stringInSlice(*lr.NodeLifecycle, state.NodeLifeCycles) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("Nodelifecyle must be [normal, spot]")})
			return
		}
	} else {
		lr.NodeLifecycle = &state.DefaultLifecycle
	}

	vars := mux.Vars(r)
	run, err := ep.executionService.CreateByAlias(
		vars["alias"],
		lr.ClusterName,
		lr.Env,
		lr.RunTags.OwnerID,
		lr.Command,
		lr.Memory,
		lr.Cpu,
		lr.Engine,
		lr.EphemeralStorage,
		lr.NodeLifecycle)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating run alias",
			"operation", "CreateRunByAlias",
			"error", fmt.Sprintf("%+v", err),
			"alias", vars["alias"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

func (ep *endpoints) StopRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	err := ep.executionService.Terminate(vars["run_id"])
	if err != nil {
		ep.logger.Log(
			"message", "problem stopping run",
			"operation", "StopRun",
			"error", fmt.Sprintf("%+v", err),
			"run_id", vars["run_id"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, map[string]bool{"terminated": true})
	}
}

func (ep *endpoints) UpdateRun(w http.ResponseWriter, r *http.Request) {
	var run state.Run
	err := ep.decodeRequest(r, &run)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	vars := mux.Vars(r)
	err = ep.executionService.UpdateStatus(vars["run_id"], run.Status, run.ExitCode)
	if err != nil {
		ep.logger.Log(
			"message", "problem updating run",
			"operation", "UpdateRun",
			"error", fmt.Sprintf("%+v", err),
			"run_id", vars["run_id"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, map[string]bool{"updated": true})
	}
}

func (ep *endpoints) GetEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := ep.executionService.Get(vars["run_id"])

	if err != nil {
		ep.logger.Log(
			"message", "problem getting run",
			"operation", "GetRun",
			"error", fmt.Sprintf("%+v", err),
			"run_id", vars["run_id"])
		ep.encodeError(w, err)
		return
	}

	if *run.Engine == state.EKSEngine {
		events, err := ep.executionService.GetEvents(run)

		if err != nil {
			_ = ep.logger.Log("message", "problem getting events")
			ep.encodeError(w, err)
			return
		}
		ep.encodeResponse(w, events)
	}
}

func (ep *endpoints) GetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := r.URL.Query()

	lastSeen := ep.getURLParam(params, "last_seen", "")
	run, err := ep.executionService.Get(vars["run_id"])

	if err != nil {
		ep.logger.Log(
			"message", "problem getting run",
			"operation", "GetRun",
			"error", fmt.Sprintf("%+v", err),
			"run_id", vars["run_id"])
		ep.encodeError(w, err)
		return
	}

	if run.Engine == nil {
		run.Engine = &state.DefaultEngine
	}

	if *run.Engine == state.ECSEngine {
		logs, newLastSeen, err := ep.ecsLogService.Logs(vars["run_id"], &lastSeen)
		if err != nil {
			ep.logger.Log(
				"message", "problem getting logs",
				"operation", "GetLogs",
				"error", fmt.Sprintf("%+v", err),
				"run_id", vars["run_id"],
				"last_seen", lastSeen)
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

	if *run.Engine == state.EKSEngine {
		logs, newLastSeen, err := ep.eksLogService.Logs(vars["run_id"], &lastSeen)
		if err != nil {
			ep.logger.Log(
				"message", "problem getting logs",
				"operation", "GetLogs",
				"error", fmt.Sprintf("%+v", err),
				"run_id", vars["run_id"],
				"last_seen", lastSeen)
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

}

func (ep *endpoints) GetGroups(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	var name string
	if len(lr.filters["name"]) > 0 {
		name = lr.filters["name"][0]
	}

	groups, err := ep.definitionService.ListGroups(lr.limit, lr.offset, &name)
	if err != nil {
		ep.logger.Log(
			"message", "problem getting groups",
			"operation", "GetGroups",
			"error", fmt.Sprintf("%+v", err))
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
		ep.logger.Log(
			"message", "problem getting tags",
			"operation", "GetTags",
			"error", fmt.Sprintf("%+v", err))
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
		ep.logger.Log(
			"message", "problem listing clusters",
			"operation", "ListClusters",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["clusters"] = clusters
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) ListWorkers(w http.ResponseWriter, r *http.Request) {
	wl, err := ep.workerService.List(state.ECSEngine)
	wlEKS, errEKS := ep.workerService.List(state.EKSEngine)

	if wl.Workers == nil {
		wl.Workers = []state.Worker{}
	}
	
	if wlEKS.Workers == nil {
		wlEKS.Workers = []state.Worker{}
	}

	if err != nil || errEKS != nil {
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["total"] = wl.Total + wlEKS.Total
		response["workers"] = append(wl.Workers, wlEKS.Workers...)
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) GetWorker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	worker, err := ep.workerService.Get(vars["worker_type"], state.DefaultEngine)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, worker)
	}
}

func (ep *endpoints) UpdateWorker(w http.ResponseWriter, r *http.Request) {
	var worker state.Worker
	err := ep.decodeRequest(r, &worker)

	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	vars := mux.Vars(r)
	updated, err := ep.workerService.Update(vars["worker_type"], worker)

	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, updated)
	}
}

func (ep *endpoints) BatchUpdateWorkers(w http.ResponseWriter, r *http.Request) {
	var wks []state.Worker
	err := ep.decodeRequest(r, &wks)

	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	updated, err := ep.workerService.BatchUpdate(wks)

	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, updated)
	}
}
