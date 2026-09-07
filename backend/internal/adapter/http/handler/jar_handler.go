// Package handler provides HTTP adapters for the application.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

type JarService interface {
	Create(ctx context.Context, name string, allocationType jar.AllocationType, allocationValue int64) (*jar.Jar, error)

	List(ctx context.Context) ([]*jar.Jar, error)

	Update(ctx context.Context, id uuid.UUID, name *string, isArchived *bool) error

	UpdateAllocations(ctx context.Context, updates []jar.JarAllocationUpdate) error
}

type JarHandler struct {
	service  JarService
	validate *validator.Validate
	logger   *slog.Logger
}

func NewJarHandler(service JarService, logger *slog.Logger) *JarHandler {
	return &JarHandler{
		service:  service,
		validate: validator.New(),
		logger:   logger,
	}
}

type CreateJarRequest struct {
	Name            string             `json:"name" validate:"required,max=100"`
	AllocationType  jar.AllocationType `json:"allocation_type" validate:"required"`
	AllocationValue int64              `json:"allocation_value"`
}

type PatchJarRequest struct {
	Name       *string `json:"name" validate:"omitempty,min=1,max=100"`
	IsArchived *bool   `json:"is_archived"`
}

type UpdateJarAllocationRequest struct {
	ID              uuid.UUID          `json:"id" validate:"required"`
	AllocationType  jar.AllocationType `json:"allocation_type" validate:"required"`
	AllocationValue int64              `json:"allocation_value"`
}

type UpdateJarAllocationsRequest struct {
	Allocations []UpdateJarAllocationRequest `json:"allocations" validate:"required,min=1"`
}

func NewJarResponse(j *jar.Jar) JarResponse {
	return JarResponse{
		ID:              j.ID,
		Name:            j.Name,
		AllocationType:  j.AllocationType,
		AllocationValue: j.AllocationValue,
		IsArchived:      j.IsArchived,
		CreatedAt:       j.CreatedAt,
		UpdatedAt:       j.UpdatedAt,
	}
}

// Create handles POST /jars.
func (h *JarHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(
		r.Context(),
		"jar create request received",
	)

	var req CreateJarRequest

	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"jar validation failed",
			slog.String("error", err.Error()),
		)

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Invalid jar request",
			"ERR_INVALID_INPUT",
		)
		return
	}

	if !req.AllocationType.IsValid() {
		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Invalid allocation type",
			"ERR_INVALID_ALLOCATION_TYPE",
		)
		return
	}

	j, err := h.service.Create(
		r.Context(),
		req.Name,
		req.AllocationType,
		req.AllocationValue,
	)
	if err != nil {
		switch {
		case errors.Is(err, jar.ErrJarNameExists):
			h.logger.WarnContext(
				r.Context(),
				"duplicate jar name attempted",
				slog.String("name", req.Name),
			)

			writeErrorJSON(
				w,
				http.StatusConflict,
				"CONFLICT",
				"Jar name already exists",
				"ERR_JAR_EXISTS",
			)

		case errors.Is(err, jar.ErrInvalidName):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid jar name",
				"ERR_INVALID_NAME",
			)

		case errors.Is(err, jar.ErrInvalidAllocationType):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid allocation type",
				"ERR_INVALID_ALLOCATION_TYPE",
			)

		case errors.Is(err, jar.ErrInvalidAllocationVal):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid allocation value",
				"ERR_INVALID_ALLOCATION_VALUE",
			)

		case errors.Is(err, jar.ErrEmptyAllocationConfiguration),
			errors.Is(err, jar.ErrNoRemainder),
			errors.Is(err, jar.ErrMultipleRemainders),
			errors.Is(err, jar.ErrPercentageExceedsLimit):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				err.Error(),
				"ERR_INVALID_ALLOCATION",
			)

		default:
			h.logger.ErrorContext(
				r.Context(),
				"failed to create jar",
				slog.String("error", err.Error()),
			)

			writeErrorJSON(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Failed to create jar",
				"ERR_INTERNAL_SERVER",
			)
		}

		return
	}

	h.logger.InfoContext(
		r.Context(),
		"jar created successfully",
		slog.String("id", j.ID.String()),
	)

	writeJSON(
		w,
		http.StatusCreated,
		NewJarResponse(j),
	)
}

// List handles GET /jars.
func (h *JarHandler) List(w http.ResponseWriter, r *http.Request) {
	jars, err := h.service.List(r.Context())
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to list jars",
			slog.String("error", err.Error()),
		)

		writeErrorJSON(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Failed to list jars",
			"ERR_INTERNAL_SERVER",
		)
		return
	}

	response := make([]JarResponse, 0, len(jars))

	for _, j := range jars {
		response = append(response, NewJarResponse(j))
	}

	writeJSON(w, http.StatusOK, response)
}

// Patch handles PATCH /jars/{id}.
func (h *JarHandler) Patch(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")

	id, err := uuid.Parse(rawID)
	if err != nil {
		h.logger.WarnContext(
			r.Context(),
			"invalid jar id format in patch",
			slog.String("raw_id", rawID),
		)

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid jar ID format",
			"ERR_INVALID_ID",
		)
		return
	}

	var req PatchJarRequest

	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Name == nil && req.IsArchived == nil {
		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"At least one field must be provided",
			"ERR_NO_FIELDS",
		)
		return
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		req.Name = &trimmedName
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"jar patch validation failed",
			slog.String("error", err.Error()),
		)

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Invalid patch format inputs",
			"ERR_INVALID_INPUT",
		)
		return
	}

	if err := h.service.Update(
		r.Context(),
		id,
		req.Name,
		req.IsArchived,
	); err != nil {
		switch {
		case errors.Is(err, jar.ErrJarNotFound):
			writeErrorJSON(
				w,
				http.StatusNotFound,
				"NOT_FOUND",
				"Jar not found",
				"ERR_JAR_NOT_FOUND",
			)

		case errors.Is(err, jar.ErrJarNameExists):
			writeErrorJSON(
				w,
				http.StatusConflict,
				"CONFLICT",
				"Jar name already exists",
				"ERR_JAR_EXISTS",
			)

		case errors.Is(err, jar.ErrInvalidName):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid jar name",
				"ERR_INVALID_NAME",
			)

		case errors.Is(err, jar.ErrEmptyAllocationConfiguration),
			errors.Is(err, jar.ErrNoRemainder),
			errors.Is(err, jar.ErrMultipleRemainders),
			errors.Is(err, jar.ErrPercentageExceedsLimit):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				err.Error(),
				"ERR_INVALID_ALLOCATION",
			)

		default:
			h.logger.ErrorContext(
				r.Context(),
				"failed to update jar",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)

			writeErrorJSON(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Failed to update jar",
				"ERR_INTERNAL_SERVER",
			)
		}

		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Message: "Jar updated successfully",
	})
}

// UpdateAllocations handles PATCH /jars/allocations.
func (h *JarHandler) UpdateAllocations(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(
		r.Context(),
		"jar allocation update request received",
	)

	var req UpdateJarAllocationsRequest

	if !decodeJSONBody(w, r, &req) {
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"jar allocation update validation failed",
			slog.String("error", err.Error()),
		)

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Invalid allocation update request",
			"ERR_INVALID_INPUT",
		)
		return
	}

	updates := make([]jar.JarAllocationUpdate, 0, len(req.Allocations))

	for _, allocation := range req.Allocations {
		if !allocation.AllocationType.IsValid() {
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid allocation type",
				"ERR_INVALID_ALLOCATION_TYPE",
			)
			return
		}

		updates = append(updates, jar.JarAllocationUpdate{
			ID:              allocation.ID,
			AllocationType:  allocation.AllocationType,
			AllocationValue: allocation.AllocationValue,
		})
	}

	if err := h.service.UpdateAllocations(
		r.Context(),
		updates,
	); err != nil {
		switch {
		case errors.Is(err, jar.ErrJarNotFound):
			writeErrorJSON(
				w,
				http.StatusNotFound,
				"NOT_FOUND",
				"Jar not found",
				"ERR_JAR_NOT_FOUND",
			)

		case errors.Is(err, jar.ErrInvalidAllocationType):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid allocation type",
				"ERR_INVALID_ALLOCATION_TYPE",
			)

		case errors.Is(err, jar.ErrInvalidAllocationVal):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				"Invalid allocation value",
				"ERR_INVALID_ALLOCATION_VALUE",
			)

		case errors.Is(err, jar.ErrEmptyAllocationConfiguration),
			errors.Is(err, jar.ErrNoRemainder),
			errors.Is(err, jar.ErrMultipleRemainders),
			errors.Is(err, jar.ErrPercentageExceedsLimit):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"VALIDATION_ERROR",
				err.Error(),
				"ERR_INVALID_ALLOCATION",
			)

		default:
			h.logger.ErrorContext(
				r.Context(),
				"failed to update jar allocations",
				slog.String("error", err.Error()),
			)

			writeErrorJSON(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"Failed to update jar allocations",
				"ERR_INTERNAL_SERVER",
			)
		}

		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{
		Message: "Jar allocations updated successfully",
	})
}

// decodeJSONBody decodes a JSON request body while enforcing the maximum
// request body size and rejecting unknown fields.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var maxBytesErr *http.MaxBytesError

		if errors.As(err, &maxBytesErr) {
			writeErrorJSON(
				w,
				http.StatusRequestEntityTooLarge,
				"REQUEST_TOO_LARGE",
				"Request body exceeds the maximum allowed size of 1 MiB",
				"ERR_BODY_TOO_LARGE",
			)
			return false
		}

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Request body contains invalid JSON",
			"ERR_INVALID_JSON",
		)
		return false
	}

	if decoder.Decode(&struct{}{}) != io.EOF {
		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Request body must contain a single JSON object",
			"ERR_INVALID_JSON",
		)
		return false
	}

	return true
}
