package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

type TemplateService interface {
	Create(context.Context, string, transaction.TransactionType, transaction.TransactionCategory, *uuid.UUID, *uuid.UUID, *uuid.UUID, *int64) (*transaction.Template, error)
	List(context.Context, ports.ListOptions) ([]*transaction.Template, error)
	GetByID(context.Context, uuid.UUID) (*transaction.Template, error)
	Update(context.Context, *transaction.Template) error
	Delete(context.Context, uuid.UUID) error
}
type TemplateHandler struct {
	service TemplateService
	logger  *slog.Logger
}

func NewTemplateHandler(service TemplateService, logger *slog.Logger) *TemplateHandler {
	return &TemplateHandler{service: service, logger: logger}
}

type TemplateRequest struct {
	Name          string                          `json:"name"`
	Type          transaction.TransactionType     `json:"type"`
	Category      transaction.TransactionCategory `json:"category"`
	FromAccountID *string                         `json:"from_account_id"`
	ToAccountID   *string                         `json:"to_account_id"`
	JarID         *string                         `json:"jar_id"`
	Amount        *int64                          `json:"amount"`
	IsArchived    *bool                           `json:"is_archived"`
}
type TemplateResponse struct {
	ID            uuid.UUID                       `json:"id"`
	Name          string                          `json:"name"`
	Type          transaction.TransactionType     `json:"type"`
	Category      transaction.TransactionCategory `json:"category"`
	FromAccountID *uuid.UUID                      `json:"from_account_id"`
	ToAccountID   *uuid.UUID                      `json:"to_account_id"`
	JarID         *uuid.UUID                      `json:"jar_id"`
	Amount        *int64                          `json:"amount"`
	IsArchived    bool                            `json:"is_archived"`
	CreatedAt     time.Time                       `json:"created_at"`
	UpdatedAt     time.Time                       `json:"updated_at"`
}

func templateResponse(t *transaction.Template) TemplateResponse {
	return TemplateResponse{ID: t.ID, Name: t.Name, Type: t.Type, Category: t.Category, FromAccountID: t.FromAccountID, ToAccountID: t.ToAccountID, JarID: t.JarID, Amount: t.Amount, IsArchived: t.IsArchived, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
}
func parseTemplateUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
func parseTemplateID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid template ID format", "ERR_INVALID_ID")
		return uuid.Nil, false
	}
	return id, true
}
func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req TemplateRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}
	from, err := parseTemplateUUID(req.FromAccountID)
	if err != nil {
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid from_account_id", "ERR_INVALID_INPUT")
		return
	}
	to, err := parseTemplateUUID(req.ToAccountID)
	if err != nil {
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid to_account_id", "ERR_INVALID_INPUT")
		return
	}
	jar, err := parseTemplateUUID(req.JarID)
	if err != nil {
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid jar_id", "ERR_INVALID_INPUT")
		return
	}
	t, err := h.service.Create(r.Context(), req.Name, req.Type, req.Category, from, to, jar, req.Amount)
	if err != nil {
		h.writeTemplateError(w, err, "create")
		return
	}
	writeJSON(w, http.StatusCreated, templateResponse(t))
}
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	archived := false
	if raw := r.URL.Query().Get("is_archived"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			writeErrorJSON(w, 400, "VALIDATION_ERROR", "is_archived must be true or false", "ERR_INVALID_INPUT")
			return
		}
		archived = parsed
	}
	limit, offset := 100, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			writeErrorJSON(w, 400, "VALIDATION_ERROR", "limit must be a positive integer", "ERR_INVALID_INPUT")
			return
		}
		limit = n
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			writeErrorJSON(w, 400, "VALIDATION_ERROR", "offset must be a non-negative integer", "ERR_INVALID_INPUT")
			return
		}
		offset = n
	}
	items, err := h.service.List(r.Context(), ports.ListOptions{Limit: limit, Offset: offset, IsArchived: archived})
	if err != nil {
		writeErrorJSON(w, 500, "INTERNAL_ERROR", "Failed to list templates", "ERR_INTERNAL_SERVER")
		return
	}
	out := make([]TemplateResponse, 0, len(items))
	for _, t := range items {
		out = append(out, templateResponse(t))
	}
	writeJSON(w, 200, out)
}
func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTemplateID(w, r)
	if !ok {
		return
	}
	t, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeTemplateError(w, err, "get")
		return
	}
	writeJSON(w, 200, templateResponse(t))
}
func (h *TemplateHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTemplateID(w, r)
	if !ok {
		return
	}
	var fields map[string]json.RawMessage
	if !decodeJSONBody(w, r, &fields) {
		return
	}
	if len(fields) == 0 {
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "At least one field must be provided", "ERR_NO_FIELDS")
		return
	}
	t, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeTemplateError(w, err, "get")
		return
	}
	decode := func(key string, dst any) bool {
		raw, exists := fields[key]
		if !exists {
			return true
		}
		if err := json.Unmarshal(raw, dst); err != nil {
			writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid "+key, "ERR_INVALID_INPUT")
			return false
		}
		return true
	}
	if !decode("name", &t.Name) || !decode("type", &t.Type) || !decode("category", &t.Category) || !decode("is_archived", &t.IsArchived) {
		return
	}
	if raw, exists := fields["from_account_id"]; exists {
		if string(raw) == "null" {
			t.FromAccountID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid from_account_id", "ERR_INVALID_INPUT")
				return
			}
			id, e := uuid.Parse(s)
			if e != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid from_account_id", "ERR_INVALID_INPUT")
				return
			}
			t.FromAccountID = &id
		}
	}
	if raw, exists := fields["to_account_id"]; exists {
		if string(raw) == "null" {
			t.ToAccountID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid to_account_id", "ERR_INVALID_INPUT")
				return
			}
			id, e := uuid.Parse(s)
			if e != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid to_account_id", "ERR_INVALID_INPUT")
				return
			}
			t.ToAccountID = &id
		}
	}
	if raw, exists := fields["jar_id"]; exists {
		if string(raw) == "null" {
			t.JarID = nil
		} else {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid jar_id", "ERR_INVALID_INPUT")
				return
			}
			id, e := uuid.Parse(s)
			if e != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid jar_id", "ERR_INVALID_INPUT")
				return
			}
			t.JarID = &id
		}
	}
	if raw, exists := fields["amount"]; exists {
		if string(raw) == "null" {
			t.Amount = nil
		} else {
			var n int64
			if err := json.Unmarshal(raw, &n); err != nil {
				writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid amount", "ERR_INVALID_INPUT")
				return
			}
			t.Amount = &n
		}
	}
	if err := h.service.Update(r.Context(), t); err != nil {
		h.writeTemplateError(w, err, "update")
		return
	}
	updated, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeTemplateError(w, err, "get")
		return
	}
	writeJSON(w, 200, templateResponse(updated))
}
func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTemplateID(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeTemplateError(w, err, "delete")
		return
	}
	writeJSON(w, 200, SuccessResponse{Message: "Template deleted successfully"})
}
func (h *TemplateHandler) writeTemplateError(w http.ResponseWriter, err error, operation string) {
	switch {
	case errors.Is(err, transaction.ErrTemplateNotFound):
		writeErrorJSON(w, 404, "NOT_FOUND", "Template not found", "ERR_TEMPLATE_NOT_FOUND")
	case errors.Is(err, transaction.ErrTemplateNameExists):
		writeErrorJSON(w, 409, "CONFLICT", "Template name already exists", "ERR_TEMPLATE_EXISTS")
	case errors.Is(err, transaction.ErrTemplateInvalidName):
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Template name cannot be empty", "ERR_INVALID_NAME")
	case errors.Is(err, transaction.ErrTemplateInvalidAmount):
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Template amount must be greater than zero", "ERR_INVALID_AMOUNT")
	case errors.Is(err, transaction.ErrInvalidCategory):
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid template category", "ERR_INVALID_CATEGORY")
	case errors.Is(err, transaction.ErrInvalidTransactionType):
		writeErrorJSON(w, 400, "VALIDATION_ERROR", "Invalid template type", "ERR_INVALID_TYPE")
	default:
		h.logger.Error("template operation failed", slog.String("operation", operation), slog.String("error", err.Error()))
		writeErrorJSON(w, 500, "INTERNAL_ERROR", "Failed to "+operation+" template", "ERR_INTERNAL_SERVER")
	}
}
