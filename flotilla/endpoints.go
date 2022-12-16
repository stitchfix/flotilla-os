package flotilla

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
	"github.com/stitchfix/flotilla-os/exceptions"
	flotillaLog "github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/utils"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type endpoints struct {
	executionService  services.ExecutionService
	definitionService services.DefinitionService
	templateService   services.TemplateService
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
	ClusterName *string        `json:"cluster,omitempty"`
	Env         *state.EnvList `json:"env,omitempty"`
}

type LaunchRequestV2 struct {
	RunTags               RunTags `json:"run_tags"`
	Command               *string
	Memory                *int64
	Cpu                   *int64
	Gpu                   *int64
	Engine                *string
	NodeLifecycle         *string               `json:"node_lifecycle"`
	ActiveDeadlineSeconds *int64                `json:"active_deadline_seconds,omitempty"`
	SparkExtension        *state.SparkExtension `json:"spark_extension,omitempty"`
	ClusterName           *string               `json:"cluster,omitempty"`
	Env                   *state.EnvList        `json:"env,omitempty"`
	Description           *string               `json:"description,omitempty"`
	CommandHash           *string               `json:"command_hash,omitempty"`
	IdempotenceKey        *string               `json:"idempotence_key,omitempty"`
	Arch                  *string               `json:"arch,omitempty"`
	Labels                *state.Labels         `json:"labels,omitempty"`
}

// RunTags represents which user is responsible for a task run
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

// Note: the difference between this method and `decodeListRequest` is that
// this method does not assume that all entities can be sorted by `group_name`.
// Instead, it relies on the IOrderable interface's DefaultOrderField method.
func (ep *endpoints) decodeOrderableListRequest(r *http.Request, orderable state.IOrderable) listRequest {
	var lr listRequest
	params := r.URL.Query()

	lr.limit, _ = strconv.Atoi(ep.getURLParam(params, "limit", "1024"))
	lr.offset, _ = strconv.Atoi(ep.getURLParam(params, "offset", "0"))
	lr.sortBy = ep.getURLParam(params, "sort_by", orderable.DefaultOrderField())
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
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func (ep *endpoints) encodeResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
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

// Fetches definition from DB using definition id.
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

// Fetches definition from DB using definition alias.
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

// Creates new definition.
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

// Updates existing definition.
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

// Deletes a defiition.
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

// List all runs, supports filtering based on environment variables.
// ListRequest is object used here to construct the query.
func (ep *endpoints) ListRuns(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)
	runList, err := ep.executionService.List(lr.limit, lr.offset, lr.order, lr.sortBy, lr.filters, lr.envFilters)
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

// List runs for a definition ID.
func (ep *endpoints) ListDefinitionRuns(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	vars := mux.Vars(r)
	definitionID, ok := vars["definition_id"]
	if ok {
		lr.filters["definition_id"] = []string{definitionID}
	}

	runList, err := ep.executionService.List(lr.limit, lr.offset, lr.order, lr.sortBy, lr.filters, lr.envFilters)
	if err != nil {
		ep.logger.Log(
			"message", "problem listing definition runs",
			"operation", "ListDefinitionRuns",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		response := ep.createListRunsResponse(runList, lr)
		ep.encodeResponse(w, response)
	}
}

// List runs based on a template id.
func (ep *endpoints) ListTemplateRuns(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	vars := mux.Vars(r)
	tplID, ok := vars["template_id"]
	if ok {
		lr.filters["executable_id"] = []string{tplID}
	}

	runList, err := ep.executionService.List(lr.limit, lr.offset, lr.order, lr.sortBy, lr.filters, lr.envFilters)
	if err != nil {
		ep.logger.Log(
			"message", "problem listing runs for template",
			"operation", "ListTemplateRuns",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		response := ep.createListRunsResponse(runList, lr)
		ep.encodeResponse(w, response)
	}
}

func (ep *endpoints) createListRunsResponse(runList state.RunList, req listRequest) map[string]interface{} {
	response := make(map[string]interface{})
	response["total"] = runList.Total
	response["history"] = runList.Runs
	response["limit"] = req.limit
	response["offset"] = req.offset
	response["sort_by"] = req.sortBy
	response["order"] = req.order
	response["env_filters"] = req.envFilters
	for k, v := range req.filters {
		response[k] = v
	}
	return response
}

// Fetches a run based on Run ID.
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

// Fetches a run based on Run ID.
func (ep *endpoints) GetPayload(w http.ResponseWriter, r *http.Request) {
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
		if run.ExecutionRequestCustom != nil {
			ep.encodeResponse(w, run.ExecutionRequestCustom)
		} else {
			ep.encodeResponse(w, map[string]string{})
		}
	}
}

// Creates a new Run (deprecated). Only present for legacy support.
func (ep *endpoints) CreateRun(w http.ResponseWriter, r *http.Request) {
	var lr LaunchRequest
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	vars := mux.Vars(r)
	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Env:              lr.Env,
			OwnerID:          "v1-unknown",
			Command:          nil,
			Memory:           nil,
			Cpu:              nil,
			Gpu:              nil,
			Engine:           &state.DefaultEngine,
			EphemeralStorage: nil,
			NodeLifecycle:    nil,
			CommandHash:      nil,
		},
	}
	run, err := ep.executionService.CreateDefinitionRunByDefinitionID(vars["definition_id"], &req)
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

// Creates a new Run (deprecated). Only present for legacy support.
func (ep *endpoints) CreateRunV2(w http.ResponseWriter, r *http.Request) {
	var lr LaunchRequestV2
	err := ep.decodeRequest(r, &lr)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	// check if OnwerEmail is present in lr.EventLabels

	if len(lr.RunTags.OwnerEmail) == 0 || len(lr.RunTags.TeamName) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("run_tags must exist in body and contain [owner_email] and [team_name]")})
		return
	}

	vars := mux.Vars(r)
	if lr.Engine == nil {
		if lr.SparkExtension != nil {
			lr.Engine = &state.EKSSparkEngine
		} else {
			lr.Engine = &state.EKSEngine
		}
	}

	if lr.CommandHash == nil && lr.Description != nil {
		lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
	}

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Env:              lr.Env,
			OwnerID:          lr.RunTags.OwnerEmail,
			Command:          nil,
			Memory:           nil,
			Cpu:              nil,
			Gpu:              nil,
			Engine:           lr.Engine,
			EphemeralStorage: nil,
			NodeLifecycle:    nil,
			SparkExtension:   lr.SparkExtension,
			Description:      lr.Description,
			CommandHash:      lr.CommandHash,
			IdempotenceKey:   lr.IdempotenceKey,
			Arch:             lr.Arch,
			Labels:           lr.Labels,
		},
	}
	run, err := ep.executionService.CreateDefinitionRunByDefinitionID(vars["definition_id"], &req)
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

// Creates a new Run.
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

	if lr.Engine == nil || *lr.Engine == "ecs" {
		if lr.SparkExtension != nil {
			lr.Engine = &state.EKSSparkEngine
		} else {
			lr.Engine = &state.EKSEngine
		}
	}

	if lr.CommandHash == nil && lr.Description != nil {
		lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
	}

	if lr.NodeLifecycle != nil {
		if !utils.StringSliceContains(state.NodeLifeCycles, *lr.NodeLifecycle) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("Nodelifecyle must be [normal, spot]")})
			return
		}
	} else {
		lr.NodeLifecycle = &state.DefaultLifecycle
	}
	vars := mux.Vars(r)

	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Env:                   lr.Env,
			OwnerID:               lr.RunTags.OwnerID,
			Command:               lr.Command,
			Memory:                lr.Memory,
			Cpu:                   lr.Cpu,
			Gpu:                   lr.Gpu,
			Engine:                lr.Engine,
			NodeLifecycle:         lr.NodeLifecycle,
			ActiveDeadlineSeconds: lr.ActiveDeadlineSeconds,
			SparkExtension:        lr.SparkExtension,
			Description:           lr.Description,
			CommandHash:           lr.CommandHash,
			IdempotenceKey:        lr.IdempotenceKey,
			Arch:                  lr.Arch,
			Labels:                lr.Labels,
		},
	}

	run, err := ep.executionService.CreateDefinitionRunByDefinitionID(vars["definition_id"], &req)

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

// Creates a new Run based on definition alias.
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

	if lr.Engine == nil || *lr.Engine == "ecs" {
		if lr.SparkExtension != nil {
			lr.Engine = &state.EKSSparkEngine
		} else {
			lr.Engine = &state.EKSEngine
		}
	}

	if lr.CommandHash == nil && lr.Description != nil {
		lr.CommandHash = aws.String(fmt.Sprintf("%x", md5.Sum([]byte(*lr.Description))))
	}

	if lr.NodeLifecycle != nil {
		if !utils.StringSliceContains(state.NodeLifeCycles, *lr.NodeLifecycle) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("Nodelifecyle must be [normal, spot]")})
			return
		}
	} else {
		lr.NodeLifecycle = &state.DefaultLifecycle
	}

	vars := mux.Vars(r)
	req := state.DefinitionExecutionRequest{
		ExecutionRequestCommon: &state.ExecutionRequestCommon{
			Env:                   lr.Env,
			OwnerID:               lr.RunTags.OwnerID,
			Command:               lr.Command,
			Memory:                lr.Memory,
			Cpu:                   lr.Cpu,
			Gpu:                   lr.Gpu,
			Engine:                lr.Engine,
			NodeLifecycle:         lr.NodeLifecycle,
			ActiveDeadlineSeconds: lr.ActiveDeadlineSeconds,
			SparkExtension:        lr.SparkExtension,
			Description:           lr.Description,
			CommandHash:           lr.CommandHash,
			IdempotenceKey:        lr.IdempotenceKey,
			Arch:                  lr.Arch,
			Labels:                lr.Labels,
		},
	}
	run, err := ep.executionService.CreateDefinitionRunByAlias(vars["alias"], &req)
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

// Stops a run based on run ID.
func (ep *endpoints) StopRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userInfo := ep.ExtractUserInfo(r)
	err := ep.executionService.Terminate(vars["run_id"], userInfo)
	if err != nil {
		ep.logger.Log(
			"message", "problem stopping run",
			"operation", "StopRun",
			"error", fmt.Sprintf("%+v", err),
			"run_id", vars["run_id"])
	}
	ep.encodeResponse(w, map[string]bool{"terminated": true})
}

// Extracts user info if present in the headers.s
func (ep *endpoints) ExtractUserInfo(r *http.Request) state.UserInfo {
	var userInfo state.UserInfo
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {

			if strings.Contains(name, "-name") {
				userInfo.Name = h
			}

			if strings.Contains(name, "-email") {
				userInfo.Email = h
			}
		}
	}
	return userInfo
}

// Update an existing run.
func (ep *endpoints) UpdateRun(w http.ResponseWriter, r *http.Request) {
	var run state.Run
	err := ep.decodeRequest(r, &run)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	vars := mux.Vars(r)
	err = ep.executionService.UpdateStatus(vars["run_id"], run.Status, run.ExitCode, run.RunExceptions, run.ExitReason)
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

// Get Pod Events (EKS only) for a run ID.
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
	var podEventList state.PodEventList
	if run.PodEvents != nil {
		podEventList.Total = len(*run.PodEvents)
		podEventList.PodEvents = *run.PodEvents
	}
	ep.encodeResponse(w, podEventList)

}

// Get logs for a run.
func (ep *endpoints) GetLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	params := r.URL.Query()

	lastSeen := ep.getURLParam(params, "last_seen", "")
	rawText := ep.getStringBoolVal(ep.getURLParam(params, "raw_text", ""))
	run, err := ep.executionService.Get(vars["run_id"])
	role := ep.getURLParam(params, "role", "driver")
	facility := ep.getURLParam(params, "facility", "stderr")

	if err != nil {
		_ = ep.logger.Log(
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

	if rawText == true {
		_ = ep.eksLogService.LogsText(vars["run_id"], w)
	} else {
		log, newLastSeen, err := ep.eksLogService.Logs(vars["run_id"], &lastSeen, &role, &facility)

		res := map[string]string{
			"log":       "",
			"last_seen": lastSeen,
		}

		if err == nil {
			res = map[string]string{
				"log":       log,
				"last_seen": *newLastSeen,
			}
		}

		ep.encodeResponse(w, res)
	}
}

// Get list of groups.
func (ep *endpoints) GetGroups(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["total"] = 0
	response["groups"] = []string{}
	ep.encodeResponse(w, response)
}

// Get listing of tags.
func (ep *endpoints) GetTags(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["total"] = 0
	response["tags"] = []string{}
	ep.encodeResponse(w, response)
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

// List active workers.
func (ep *endpoints) ListWorkers(w http.ResponseWriter, r *http.Request) {
	wl, err := ep.workerService.List(state.EKSEngine)
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

// Get information about an active worker.
func (ep *endpoints) GetWorker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	worker, err := ep.workerService.Get(vars["worker_type"], state.DefaultEngine)
	if err != nil {
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, worker)
	}
}

// Update worker counts.
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

// Update batches of workers - used to turn on/off in bulk.
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

func (ep *endpoints) getStringBoolVal(s string) bool {
	l := strings.ToLower(s)

	if l == "true" {
		return true
	}

	return false
}

// Create a new template run based on template name/alias.
func (ep *endpoints) CreateTemplateRunByName(w http.ResponseWriter, r *http.Request) {
	var req state.TemplateExecutionRequest
	err := ep.decodeRequest(r, &req)

	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	if len(req.OwnerID) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("request payload must contain [owner_id]; the run_tags field is deprecated for the v7 endpoint.")})
		return
	}

	req.Engine = &state.DefaultEngine

	if req.NodeLifecycle != nil {
		if !utils.StringSliceContains(state.NodeLifeCycles, *req.NodeLifecycle) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("Nodelifecyle must be [normal, spot]")})
			return
		}
	} else {
		req.NodeLifecycle = &state.DefaultLifecycle
	}
	vars := mux.Vars(r)

	run, err := ep.executionService.CreateTemplateRunByTemplateName(vars["template_name"], vars["template_version"], &req)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating template run",
			"operation", "CreateTemplateRun",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}

}

// Create a new template run based on template id.
func (ep *endpoints) CreateTemplateRun(w http.ResponseWriter, r *http.Request) {
	var req state.TemplateExecutionRequest
	err := ep.decodeRequest(r, &req)

	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	if len(req.OwnerID) == 0 {
		ep.encodeError(w, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("request payload must contain [owner_id]; the run_tags field is deprecated for the v7 endpoint.")})
		return
	}

	req.Engine = &state.DefaultEngine

	if req.NodeLifecycle != nil {
		if !utils.StringSliceContains(state.NodeLifeCycles, *req.NodeLifecycle) {
			ep.encodeError(w, exceptions.MalformedInput{
				ErrorString: fmt.Sprintf("Nodelifecyle must be [normal, spot]")})
			return
		}
	} else {
		req.NodeLifecycle = &state.DefaultLifecycle
	}
	vars := mux.Vars(r)

	run, err := ep.executionService.CreateTemplateRunByTemplateID(vars["template_id"], &req)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating template run",
			"operation", "CreateTemplateRun",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, run)
	}
}

// List all templates.
func (ep *endpoints) ListTemplates(w http.ResponseWriter, r *http.Request) {
	var (
		tl  state.TemplateList
		err error
	)
	lr := ep.decodeOrderableListRequest(r, &state.Template{})

	params := r.URL.Query()
	latestOnly := ep.getStringBoolVal(ep.getURLParam(params, "latest_only", "true"))

	if latestOnly == true {
		tl, err = ep.templateService.ListLatestOnly(lr.limit, lr.offset, lr.sortBy, lr.order)
	} else {
		tl, err = ep.templateService.List(lr.limit, lr.offset, lr.sortBy, lr.order)
	}

	if tl.Templates == nil {
		tl.Templates = []state.Template{}
	}
	if err != nil {
		ep.logger.Log(
			"message", "problem listing templates",
			"operation", "ListTemplates",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		response := make(map[string]interface{})
		response["total"] = tl.Total
		response["templates"] = tl.Templates
		response["limit"] = lr.limit
		response["offset"] = lr.offset
		response["sort_by"] = lr.sortBy
		response["order"] = lr.order
		ep.encodeResponse(w, response)
	}
}

// Get a template.
func (ep *endpoints) GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tpl, err := ep.templateService.GetByID(vars["template_id"])
	if err != nil {
		ep.logger.Log(
			"message", "problem getting templates",
			"operation", "GetTemplate",
			"error", fmt.Sprintf("%+v", err),
			"template_id", vars["template_id"])
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, tpl)
	}
}

// Create a template.
func (ep *endpoints) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req state.CreateTemplateRequest
	err := ep.decodeRequest(r, &req)
	if err != nil {
		ep.encodeError(w, exceptions.MalformedInput{ErrorString: err.Error()})
		return
	}

	created, err := ep.templateService.Create(&req)
	if err != nil {
		ep.logger.Log(
			"message", "problem creating template",
			"operation", "CreateTemplate",
			"error", fmt.Sprintf("%+v", err))
		ep.encodeError(w, err)
	} else {
		ep.encodeResponse(w, created)
	}
}
