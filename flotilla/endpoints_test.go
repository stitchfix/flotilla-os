package flotilla

import (
	"bytes"
	"encoding/json"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/services"
	"github.com/stitchfix/flotilla-os/state"
	"github.com/stitchfix/flotilla-os/testutils"
	muxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorilla/mux"
	"net/http/httptest"
	"testing"
)

func setUp(t *testing.T) *muxtrace.Router {
	confDir := "../conf"
	c, _ := config.NewConfig(&confDir)
	imp := testutils.ImplementsAllTheThings{
		T: t,
		Definitions: map[string]state.Definition{
			"A": {DefinitionID: "A", Alias: "aliasA"},
			"B": {DefinitionID: "B", Alias: "aliasB"},
			"C": {DefinitionID: "C", Alias: "aliasC", ExecutableResources: state.ExecutableResources{Image: "invalidimage"}},
		},
		Runs: map[string]state.Run{
			"runA": {DefinitionID: "A", ClusterName: "A",
				GroupName: "A",
				RunID:     "runA", Status: state.StatusRunning},
			"runB": {DefinitionID: "B", ClusterName: "B",
				GroupName: "B", RunID: "runB",
				InstanceDNSName: "cupcakedns", InstanceID: "cupcakeid"},
		},
		Qurls: map[string]string{
			"A": "a/",
			"B": "b/",
		},
		Groups: []string{"g1", "g2", "g3"},
		Tags:   []string{"t1", "t2", "t3"},
	}
	ds, _ := services.NewDefinitionService(&imp)
	es, _ := services.NewExecutionService(c, &imp, &imp, &imp, &imp)
	ls, _ := services.NewLogService(&imp, &imp)
	ep := endpoints{definitionService: ds, executionService: es, eksLogService: ls}
	return NewRouter(ep)
}

func TestEndpoints_CreateDefinition(t *testing.T) {
	router := setUp(t)

	newDef := `{"alias":"cupcake", "memory":100, "group_name":"cupcake", "image":"someimage", "command":"echo 'hi'"}`
	req := httptest.NewRequest("POST", "/api/v1/task", bytes.NewBufferString(newDef))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Definition{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(r.DefinitionID) == 0 {
		t.Errorf("Expected non-empty definition id")
	}
}

func TestEndpoints_UpdateDefinition(t *testing.T) {
	router := setUp(t)

	updatedDef := `{"image":"updatedImage"}`
	req := httptest.NewRequest("PUT", "/api/v1/task/A", bytes.NewBufferString(updatedDef))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Definition{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if r.Image != "updatedImage" {
		t.Errorf("Expected image [updatedImage] but was [%s]", r.Image)
	}
}

func TestEndpoints_CreateRun(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}]}`
	req := httptest.NewRequest("PUT", "/api/v1/task/A/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}
}

func TestEndpoints_CreateRun2(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}], "run_tags":{"owner_email":"flotilla@github.com", "team_name":"thebest"}}`
	req := httptest.NewRequest("PUT", "/api/v2/task/A/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}

	if r.User != "flotilla@github.com" {
		t.Errorf("Expected new run to have user set to run_tags.owner_email but was [%s]", r.User)
	}
}

func TestEndpoints_CreateRun4(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}], "run_tags":{"owner_id":"flotilla"}, "labels": {"foo": "bar"}}`
	req := httptest.NewRequest("PUT", "/api/v4/task/A/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}

	if len(r.Labels) != 1 || r.Labels["foo"] != "bar" {
		labelRes, _ := json.Marshal(r.Labels)
		t.Errorf(string(labelRes))
	}

	if r.User != "flotilla" {
		t.Errorf("Expected new run to have user set to run_tags.owner_id but was [%s]", r.User)
	}
}

func TestEndpoints_CreateRunByAlias(t *testing.T) {
	router := setUp(t)

	newRun := `{"cluster":"cupcake", "env":[{"name":"E1","value":"V1"}], "run_tags":{"owner_id":"flotilla"}}`
	req := httptest.NewRequest("PUT", "/api/v1/task/alias/aliasA/execute", bytes.NewBufferString(newRun))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	r := state.Run{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if len(r.RunID) == 0 {
		t.Errorf("Expected non-empty run id")
	}

	if r.Status != state.StatusQueued {
		t.Errorf("Expected new run to have status [%s] but was [%s]", state.StatusQueued, r.Status)
	}

	if r.User != "flotilla" {
		t.Errorf("Expected new run to have user set to run_tags.owner_id but was [%s]", r.User)
	}
}

func TestEndpoints_DeleteDefinition(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("DELETE", "/api/v1/task/A", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Errorf(err.Error())
	}
	if _, ok := ack["deleted"]; !ok {
		t.Errorf("Expected [deleted] acknowledgement")
	}
}

func TestEndpoints_GetDefinition(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/task/A", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r state.Definition
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if r.DefinitionID != "A" {
		t.Errorf("Expected definition_id [A] but was [%s]", r.DefinitionID)
	}

	if r.Env == nil {
		t.Errorf("Expected non-nil environment")
	}
}

func TestEndpoints_GetDefinitionByAlias(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/task/alias/aliasA", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r state.Definition
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if r.DefinitionID != "A" {
		t.Errorf("Expected definition_id [A] but was [%s]", r.DefinitionID)
	}

	if r.Env == nil {
		t.Errorf("Expected non-nil environment")
	}
}

func TestEndpoints_GetGroups(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/groups", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["groups"]; !ok {
		t.Errorf("Expected groups in response")
	}

	groups, _ := r["groups"]
	if _, ok := groups.([]interface{}); !ok {
		t.Errorf("Cannot cast groups to list, expected list")
	}
}

func TestEndpoints_GetLogs(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/runA/logs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if _, ok := r["log"]; !ok {
		t.Errorf("Expected log in response")
	}
}

func TestEndpoints_GetRun(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/history/runA", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r state.Run
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if r.RunID != "runA" {
		t.Errorf("Expected run with runID [runA] but was [%s]", r.RunID)
	}
}

func TestEndpoints_GetRun2(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/history/runB", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var other map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&other)
	if err != nil {
		t.Errorf(err.Error())
	}

	instance, ok := other["instance"]
	if !ok {
		t.Errorf("Expected [instance] in response")
	}

	if _, ok = instance.(map[string]interface{}); !ok {
		t.Errorf("Expected [instance] in response to be a map")
	}
}

func TestEndpoints_GetTags(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/tags", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["tags"]; !ok {
		t.Errorf("Expected tags in response")
	}

	tags, _ := r["tags"]
	if _, ok := tags.([]interface{}); !ok {
		t.Errorf("Cannot cast tags to list, expected list")
	}
}

func TestEndpoints_ListDefinitions(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("GET", "/api/v1/task?limit=100&offset=2&sort_by=alias&order=desc&group_name=cupcake&env=E1%7CV1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["definitions"]; !ok {
		t.Errorf("Expected definitions in response")
	}

	if _, ok := r["limit"]; !ok {
		t.Errorf("Expected limit in response")
	}

	if _, ok := r["offset"]; !ok {
		t.Errorf("Expected offset in response")
	}

	if _, ok := r["sort_by"]; !ok {
		t.Errorf("Expected sort_by in response")
	}

	if _, ok := r["order"]; !ok {
		t.Errorf("Expected order in response")
	}

	if _, ok := r["group_name"]; !ok {
		t.Errorf("Expected [group_name] filter in response")
	}

	if _, ok := r["env_filters"]; !ok {
		t.Errorf("Expected env_filters in response")
	}

	definitions, _ := r["definitions"]
	if _, ok := definitions.([]interface{}); !ok {
		t.Errorf("Cannot cast definitions to list, expected list")
	}

	envFilters, _ := r["env_filters"]
	if _, ok := envFilters.(map[string]interface{}); !ok {
		t.Errorf("Cannot cast env_filters to map, expected map")
	}

	envFiltersMap := envFilters.(map[string]interface{})
	e1Filter, ok := envFiltersMap["E1"]
	if !ok {
		t.Errorf("Expected env_filters to contain key [E1]")
	}

	if e1Filter.(string) != "V1" {
		t.Errorf("Expected env_filter [E1:V1]")
	}
}

func TestEndpoints_ListRuns(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest(
		"GET",
		"/api/v1/history?status=RUNNING&status=QUEUED&limit=100&offset=2&sort_by=started_at&order=desc&cluster=cupcake&env=E1%7CV1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var r map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		t.Errorf(err.Error())
	}

	if _, ok := r["total"]; !ok {
		t.Errorf("Expected total in response")
	}

	if _, ok := r["history"]; !ok {
		t.Errorf("Expected runs in response")
	}

	if _, ok := r["limit"]; !ok {
		t.Errorf("Expected limit in response")
	}

	if _, ok := r["offset"]; !ok {
		t.Errorf("Expected offset in response")
	}

	if _, ok := r["sort_by"]; !ok {
		t.Errorf("Expected sort_by in response")
	}

	if _, ok := r["order"]; !ok {
		t.Errorf("Expected order in response")
	}

	if _, ok := r["cluster"]; !ok {
		t.Errorf("Expected [cluster] filter in response")
	}

	if _, ok := r["env_filters"]; !ok {
		t.Errorf("Expected env_filters in response")
	}

	if _, ok := r["status"]; !ok {
		t.Errorf("Expected [status] filter in response")
	}

	runs, _ := r["history"]
	if _, ok := runs.([]interface{}); !ok {
		t.Errorf("Cannot cast runs to list, expected list")
	}

	statusFilters, _ := r["status"]
	if _, ok := statusFilters.([]interface{}); !ok {
		t.Errorf("Cannot cast status filters to list, expected list")
	}

	expectedStatusFilters := map[string]bool{"RUNNING": true, "QUEUED": true}
	statusFiltersList := statusFilters.([]interface{})
	if len(statusFiltersList) != 2 {
		t.Errorf("Expected 2 status filters, was %v", len(statusFiltersList))
	}
	for _, statusFilter := range statusFiltersList {
		if _, ok := expectedStatusFilters[statusFilter.(string)]; !ok {
			t.Errorf("Unexpected status filter: %s", statusFilter.(string))
		}
	}

	envFilters, _ := r["env_filters"]
	if _, ok := envFilters.(map[string]interface{}); !ok {
		t.Errorf("Cannot cast env_filters to map, expected map")
	}

	envFiltersMap := envFilters.(map[string]interface{})
	e1Filter, ok := envFiltersMap["E1"]
	if !ok {
		t.Errorf("Expected env_filters to contain key [E1]")
	}

	if e1Filter.(string) != "V1" {
		t.Errorf("Expected env_filter [E1:V1]")
	}
}

func TestEndpoints_StopRun(t *testing.T) {
	router := setUp(t)

	req := httptest.NewRequest("DELETE", "/api/v1/task/A/history/runA", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	resp := w.Result()

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type [application/json; charset=utf-8], but was [%s]", resp.Header.Get("Content-Type"))
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, was %v", resp.StatusCode)
	}

	var ack map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&ack)
	if err != nil {
		t.Errorf(err.Error())
	}
	if _, ok := ack["terminated"]; !ok {
		t.Errorf("Expected [terminated] acknowledgement")
	}
}
