// Package handler provides HTTP adapters for the application.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

type AccountService interface {
	Create(ctx context.Context, name string) (*account.Account, error)
	List(ctx context.Context) ([]*account.Account, error)
	Update(ctx context.Context, id uuid.UUID, name *string, isArchived *bool) error
}

type AccountHandler struct {
	service  AccountService
	validate *validator.Validate
	logger   *slog.Logger
}

func NewAccountHandler(service AccountService, logger *slog.Logger) *AccountHandler {
	return &AccountHandler{
		service:  service,
		validate: validator.New(),
		logger:   logger,
	}
}

type CreateAccountRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

type PatchAccountRequest struct {
	Name       *string `json:"name" validate:"omitempty,min=1,max=100"`
	IsArchived *bool   `json:"is_archived"`
}

func NewAccountResponse(a *account.Account) AccountResponse {
	return AccountResponse{
		ID:         a.ID,
		Name:       a.Name,
		IsArchived: a.IsArchived,
		UpdatedAt:  a.UpdatedAt,
	}
}

// Create handles POST /accounts.
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(
		r.Context(),
		"account create request received",
	)

	var req CreateAccountRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"failed to decode account create request body",
			slog.String("error", err.Error()),
		)

		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Request body is invalid",
			Code:    "ERR_BAD_REQUEST",
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"account validation failed",
			slog.String("error", err.Error()),
		)

		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid account name",
			Code:    "ERR_INVALID_NAME",
		})
		return
	}

	a, err := h.service.Create(r.Context(), req.Name)
	if err != nil {
		switch {
		case errors.Is(err, account.ErrAccountNameExists):
			h.logger.WarnContext(
				r.Context(),
				"duplicate account name attempted",
				slog.String("name", req.Name),
			)

			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error:   "CONFLICT",
				Message: "Account name already exists",
				Code:    "ERR_ACCOUNT_EXISTS",
			})

		case errors.Is(err, account.ErrInvalidName):
			h.logger.WarnContext(
				r.Context(),
				"invalid account name from service",
				slog.String("error", err.Error()),
			)

			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "VALIDATION_ERROR",
				Message: "Invalid account name",
				Code:    "ERR_INVALID_NAME",
			})

		default:
			h.logger.ErrorContext(
				r.Context(),
				"failed to create account",
				slog.String("error", err.Error()),
			)

			writeJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: "Failed to create account",
				Code:    "ERR_INTERNAL_SERVER",
			})
		}

		return
	}

	h.logger.InfoContext(
		r.Context(),
		"account created successfully",
		slog.String("id", a.ID.String()),
	)

	writeJSON(w, http.StatusCreated, NewAccountResponse(a))
}

// List handles GET /accounts.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.service.List(r.Context())
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to list accounts",
			slog.String("error", err.Error()),
		)

		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to list accounts",
			Code:    "ERR_INTERNAL_SERVER",
		})
		return
	}

	response := make([]AccountResponse, 0, len(accounts))

	for _, a := range accounts {
		response = append(response, NewAccountResponse(a))
	}

	writeJSON(w, http.StatusOK, response)
}

// Patch handles PATCH /accounts/{id}.
func (h *AccountHandler) Patch(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")

	id, err := uuid.Parse(rawID)
	if err != nil {
		h.logger.WarnContext(
			r.Context(),
			"invalid account id format in patch",
			slog.String("raw_id", rawID),
		)

		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid account ID format",
			Code:    "ERR_INVALID_ID",
		})
		return
	}

	var req PatchAccountRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"failed to decode patch body",
			slog.String("error", err.Error()),
		)

		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Request body is invalid",
			Code:    "ERR_BAD_REQUEST",
		})
		return
	}

	if req.Name == nil && req.IsArchived == nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "At least one field must be provided",
			Code:    "ERR_NO_FIELDS",
		})
		return
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		req.Name = &trimmedName
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"account patch validation failed",
			slog.String("error", err.Error()),
		)

		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Invalid patch format inputs",
			Code:    "ERR_INVALID_INPUT",
		})
		return
	}

	if err := h.service.Update(
		r.Context(),
		id,
		req.Name,
		req.IsArchived,
	); err != nil {
		switch {
		case errors.Is(err, account.ErrAccountNotFound):
			writeJSON(w, http.StatusNotFound, ErrorResponse{
				Error:   "NOT_FOUND",
				Message: "Account not found",
				Code:    "ERR_ACCOUNT_NOT_FOUND",
			})

		case errors.Is(err, account.ErrAccountNameExists):
			writeJSON(w, http.StatusConflict, ErrorResponse{
				Error:   "CONFLICT",
				Message: "Account name already exists",
				Code:    "ERR_ACCOUNT_EXISTS",
			})

		case errors.Is(err, account.ErrInvalidName):
			writeJSON(w, http.StatusBadRequest, ErrorResponse{
				Error:   "VALIDATION_ERROR",
				Message: "Invalid account name",
				Code:    "ERR_INVALID_NAME",
			})

		default:
			h.logger.ErrorContext(
				r.Context(),
				"failed to update account",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)

			writeJSON(w, http.StatusInternalServerError, ErrorResponse{
				Error:   "INTERNAL_ERROR",
				Message: "Failed to update account",
				Code:    "ERR_INTERNAL_SERVER",
			})
		}

		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Message: "Account updated successfully",
	})
}
