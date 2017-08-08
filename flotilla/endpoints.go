package flotilla

import (
	"encoding/json"
	"github.com/stitchfix/flotilla-os/exceptions"
	"github.com/stitchfix/flotilla-os/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type endpoints struct {
	executionService  services.ExecutionService
	definitionService services.DefinitionService
}

type listRequest struct {
	limit      int
	offset     int
	sortBy     string
	order      string
	filters    map[string]string
	envFilters map[string]string
}

func (ep *endpoints) getURLParam(v url.Values, key string, defaultValue string) string {
	val, ok := v[key]
	if ok && len(val) > 0 {
		return val[0]
	} else {
		return defaultValue
	}
}

func (ep *endpoints) getFilters(params url.Values, nonFilters map[string]bool) (map[string]string, map[string]string) {
	filters := make(map[string]string)
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
				filters[k] = v[0]
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

func (ep *endpoints) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	lr := ep.decodeListRequest(r)

	definitionList, err := ep.definitionService.List(
		lr.limit, lr.offset, lr.sortBy, lr.order, lr.filters, lr.envFilters)
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

func (ep *endpoints) encodeResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
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

func (ep *endpoints) UpdateRun(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) GetLogs(w http.ResponseWriter, r *http.Request) {

}

func (ep *endpoints) GetGroups(w http.ResponseWriter, r *http.Request) {

}
