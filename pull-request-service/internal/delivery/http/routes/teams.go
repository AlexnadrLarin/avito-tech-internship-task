package routes

import (
	"pull-request-service/internal/delivery/http/handlers"

	"github.com/gorilla/mux"
)

func SetupTeamRoutes(api *mux.Router, h *handlers.TeamHandler) {
	teamsApi := api.PathPrefix("/team").Subrouter()

	teamsApi.HandleFunc("/add", h.Add).Methods("POST")
	teamsApi.HandleFunc("/get", h.Get).Methods("GET")
}
