package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type AccountResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	IsArchived bool      `json:"is_archived"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
