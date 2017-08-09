package flotilla

import "github.com/gorilla/mux"

func NewRouter(ep endpoints) *mux.Router {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.HandleFunc("/task", ep.ListDefinitions).Methods("GET")
	s.HandleFunc("/task", ep.CreateDefinition).Methods("POST")
	s.HandleFunc("/task/{definition_id}", ep.GetDefinition).Methods("GET")
	s.HandleFunc("/task/{definition_id}", ep.UpdateDefinition).Methods("PUT")
	s.HandleFunc("/task/{definition_id}", ep.DeleteDefinition).Methods("DELETE")
	s.HandleFunc("/task/{definition_id}/execute", ep.CreateRun).Methods("PUT")

	s.HandleFunc("/history", ep.ListRuns).Methods("GET")
	s.HandleFunc("/history/{run_id}", ep.GetRun).Methods("GET")
	s.HandleFunc("/task/history/{run_id}", ep.GetRun).Methods("GET")
	s.HandleFunc("/task/{definition_id}/history", ep.ListRuns).Methods("GET")
	s.HandleFunc("/task/{definition_id}/history/{run_id}", ep.GetRun).Methods("GET")
	s.HandleFunc("/task/{definition_id}/history/{run_id}", ep.StopRun).Methods("DELETE")

	s.HandleFunc("/{run_id}/status", ep.UpdateRun).Methods("PUT")
	s.HandleFunc("/{run_id}/logs", ep.GetLogs).Methods("GET")
	s.HandleFunc("/groups", ep.GetGroups).Methods("GET")
	s.HandleFunc("/tags", ep.GetTags).Methods("GET")
	return s
}
