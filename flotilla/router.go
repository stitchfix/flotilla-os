package flotilla

import "github.com/gorilla/mux"

func NewRouter(ep endpoints) *mux.Router {
	r := mux.NewRouter()
	v1 := r.PathPrefix("/api/v1").Subrouter()

	v1.HandleFunc("/task", ep.ListDefinitions).Methods("GET")
	v1.HandleFunc("/task", ep.CreateDefinition).Methods("POST")
	v1.HandleFunc("/task/{definition_id}", ep.GetDefinition).Methods("GET")
	v1.HandleFunc("/task/{definition_id}", ep.UpdateDefinition).Methods("PUT")
	v1.HandleFunc("/task/{definition_id}", ep.DeleteDefinition).Methods("DELETE")
	v1.HandleFunc("/task/{definition_id}/execute", ep.CreateRun).Methods("PUT")
	v1.HandleFunc("/task/alias/{alias}", ep.GetDefinitionByAlias).Methods("GET")
	v1.HandleFunc("/task/alias/{alias}/execute", ep.CreateRunByAlias).Methods("PUT")

	v1.HandleFunc("/history", ep.ListRuns).Methods("GET")
	v1.HandleFunc("/history/{run_id}", ep.GetRun).Methods("GET")
	v1.HandleFunc("/task/history/{run_id}", ep.GetRun).Methods("GET")
	v1.HandleFunc("/task/{definition_id}/history", ep.ListRuns).Methods("GET")
	v1.HandleFunc("/task/{definition_id}/history/{run_id}", ep.GetRun).Methods("GET")
	v1.HandleFunc("/task/{definition_id}/history/{run_id}", ep.StopRun).Methods("DELETE")

	v1.HandleFunc("/{run_id}/status", ep.UpdateRun).Methods("PUT")
	v1.HandleFunc("/{run_id}/logs", ep.GetLogs).Methods("GET")
	v1.HandleFunc("/groups", ep.GetGroups).Methods("GET")
	v1.HandleFunc("/tags", ep.GetTags).Methods("GET")

	v2 := r.PathPrefix("/api/v2").Subrouter()
	v2.HandleFunc("/task/{definition_id}/execute", ep.CreateRunV2).Methods("PUT")

	v4 := r.PathPrefix("/api/v4").Subrouter()
	v4.HandleFunc("/task/{definition_id}/execute", ep.CreateRunV4).Methods("PUT")
	return r
}
