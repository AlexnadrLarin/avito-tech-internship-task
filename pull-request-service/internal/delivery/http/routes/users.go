package routes

import (
	"pull-request-service/internal/delivery/http/handlers"

	"github.com/gorilla/mux"
)

func SetupUsersRoutes(api *mux.Router, h *handlers.UsersHandler) {
	usersApi := api.PathPrefix("/users").Subrouter()

	usersApi.HandleFunc("/setIsActive", h.SetIsActive).Methods("POST")
	usersApi.HandleFunc("/getReview", h.GetReview).Methods("GET")
	usersApi.HandleFunc("/getReviewsStats", h.GetReviewsStats).Methods("GET")

}
