package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

// ErrorResponse represents an HTTP error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// AccountResponse represents an account in an HTTP response.
type AccountResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Balance    int64     `json:"balance"`
	IsArchived bool      `json:"is_archived"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// JarResponse represents a jar in an HTTP response.
type JarResponse struct {
	ID              uuid.UUID          `json:"id"`
	Name            string             `json:"name"`
	AllocationType  jar.AllocationType `json:"allocation_type"`
	AllocationValue int64              `json:"allocation_value"`
	IsArchived      bool               `json:"is_archived"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// SuccessResponse represents a successful HTTP response.
type SuccessResponse struct {
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeErrorJSON(w http.ResponseWriter, status int, err string, message string, code string) {
	writeJSON(w, status, ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	})
}
