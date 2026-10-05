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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

type JarService interface {
	Create(ctx context.Context, name string, allocationType jar.AllocationType, allocationValue int64) (*jar.Jar, error)

	ListWithStats(
		ctx context.Context,
		startDate *time.Time,
		endDate *time.Time,
	) ([]*jar.Jar, map[uuid.UUID]int64, map[uuid.UUID]int64, error)

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

const dateLayout = "2006-01-02"

// parseDateRange reads start_date and end_date from query.
// Returns [start, end) range. End is exclusive.
func parseDateRange(r *http.Request, now time.Time) (time.Time, time.Time, error) {
	q := r.URL.Query()
	startStr := q.Get("start_date")
	endStr := q.Get("end_date")

	// no params: default to current month
	if startStr == "" && endStr == "" {
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, start.AddDate(0, 1, 0), nil
	}

	if startStr == "" || endStr == "" {
		return time.Time{}, time.Time{}, errors.New("start_date and end_date must be provided together")
	}

	start, err := time.ParseInLocation(dateLayout, startStr, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("start_date must be in YYYY-MM-DD format")
	}

	end, err := time.ParseInLocation(dateLayout, endStr, now.Location())
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("end_date must be in YYYY-MM-DD format")
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, errors.New("end_date must not be before start_date")
	}

	// limit range, stop huge queries
	if end.Sub(start) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, errors.New("date range must not exceed 366 days")
	}

	// end_date inclusive for user, exclusive for query
	return start, end.AddDate(0, 0, 1), nil
}

// List handles GET /jars?start_date=2026-10-01&end_date=2026-10-31
func (h *JarHandler) List(w http.ResponseWriter, r *http.Request) {
	startDate, endDate, err := parseDateRange(r, time.Now())
	if err != nil {
		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			err.Error(),
			"ERR_INVALID_DATE_RANGE",
		)
		return
	}

	jars, allocatedMap, usedMap, err := h.service.ListWithStats(
		r.Context(),
		&startDate,
		&endDate,
	)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"failed to list jars with statistics",
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
		jarResponse := NewJarResponse(j)

		jarResponse.AllocatedAmount = allocatedMap[j.ID]
		jarResponse.UsedAmount = usedMap[j.ID]
		jarResponse.AvailableAmount = jarResponse.AllocatedAmount - jarResponse.UsedAmount

		if jarResponse.AllocatedAmount > 0 {
			jarResponse.UsedPercentage = float64(jarResponse.UsedAmount) /
				float64(jarResponse.AllocatedAmount) * 100

			jarResponse.AvailablePercentage = float64(jarResponse.AvailableAmount) /
				float64(jarResponse.AllocatedAmount) * 100
		}

		response = append(response, jarResponse)
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
