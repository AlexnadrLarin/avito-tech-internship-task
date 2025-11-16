package helpers

import (
	"encoding/json"
	"net/http"

	"pull-request-service/internal/models"
)

func WriteError(w http.ResponseWriter, status int, code models.ErrorCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(models.ErrorResponse{
		Error: models.ErrorObject{
			Code:    code,
			Message: msg,
		},
	})
}

func WriteSuccess(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
