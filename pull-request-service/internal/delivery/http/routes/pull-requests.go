package routes

import (
	"pull-request-service/internal/delivery/http/handlers"

	"github.com/gorilla/mux"
)

func SetupPullRequestRoutes(api *mux.Router, h *handlers.PullRequestHandler) {
	pullRequestsApi := api.PathPrefix("/pullRequest").Subrouter()

	pullRequestsApi.HandleFunc("/create", h.Create).Methods("POST")
	pullRequestsApi.HandleFunc("/merge", h.Merge).Methods("POST")
	pullRequestsApi.HandleFunc("/reassign", h.Reassign).Methods("POST")
}
