package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

func SetupMainRouter(loggingMw func(http.Handler) http.Handler) *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(loggingMw)

	return api
}
