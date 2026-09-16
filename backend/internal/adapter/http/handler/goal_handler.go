package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
)

type GoalService interface {
	Create(ctx context.Context, name string, target int64, deadline time.Time) (*goal.Goal, error)
	List(ctx context.Context) ([]*goal.Goal, error)
	Update(ctx context.Context, id uuid.UUID, name *string, target *int64, deadline *time.Time, isArchived *bool) error
}

// GoalHandler handles HTTP requests for goals.
type GoalHandler struct {
	service  GoalService
	validate *validator.Validate
	logger   *slog.Logger
}

// NewGoalHandler creates a new GoalHandler.
func NewGoalHandler(service GoalService, logger *slog.Logger) *GoalHandler {
	return &GoalHandler{
		service:  service,
		validate: validator.New(),
		logger:   logger,
	}
}

type CreateGoalRequest struct {
	Name     string    `json:"name" validate:"required,max=100"`
	Target   int64     `json:"target" validate:"required,gt=0"`
	Deadline time.Time `json:"deadline" validate:"required"`
}

type PatchGoalRequest struct {
	Name       *string    `json:"name" validate:"omitempty,min=1,max=100"`
	Target     *int64     `json:"target" validate:"omitempty,gt=0"`
	Deadline   *time.Time `json:"deadline" validate:"omitempty"`
	IsArchived *bool      `json:"is_archived"`
}

type GoalResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Target     int64     `json:"target"`
	Deadline   time.Time `json:"deadline"`
	IsArchived bool      `json:"is_archived"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewGoalResponse(g *goal.Goal) GoalResponse {
	return GoalResponse{
		ID:         g.ID,
		Name:       g.Name,
		Target:     g.Target,
		Deadline:   g.Deadline,
		IsArchived: g.IsArchived,
		CreatedAt:  g.CreatedAt,
		UpdatedAt:  g.UpdatedAt,
	}
}

// Create handles POST /goals.
func (h *GoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateGoalRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if err := h.validate.Struct(req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid goal request", "ERR_INVALID_INPUT")
		return
	}

	g, err := h.service.Create(r.Context(), req.Name, req.Target, req.Deadline)
	if err != nil {
		switch {
		case errors.Is(err, goal.ErrGoalNameExists):
			writeErrorJSON(w, http.StatusConflict, "CONFLICT", "Goal name already exists", "ERR_GOAL_EXISTS")
		case errors.Is(err, goal.ErrInvalidName):
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid goal name", "ERR_INVALID_NAME")
		case errors.Is(err, goal.ErrInvalidTarget):
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid goal target", "ERR_INVALID_TARGET")
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create goal", "ERR_INTERNAL_SERVER")
		}
		return
	}
	writeJSON(w, http.StatusCreated, NewGoalResponse(g))
}

// List handles GET /goals.
func (h *GoalHandler) List(w http.ResponseWriter, r *http.Request) {
	goals, err := h.service.List(r.Context())
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list goals", "ERR_INTERNAL_SERVER")
		return
	}

	response := make([]GoalResponse, 0, len(goals))
	for _, g := range goals {
		response = append(response, NewGoalResponse(g))
	}
	writeJSON(w, http.StatusOK, response)
}

// Patch handles PATCH /goals/{id}.
func (h *GoalHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid goal ID format", "ERR_INVALID_ID")
		return
	}

	var req PatchGoalRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Name == nil && req.Target == nil && req.Deadline == nil && req.IsArchived == nil {
		writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "At least one field must be provided", "ERR_NO_FIELDS")
		return
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		req.Name = &trimmedName
	}

	if err := h.validate.Struct(req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid patch format inputs", "ERR_INVALID_INPUT")
		return
	}

	if err := h.service.Update(r.Context(), id, req.Name, req.Target, req.Deadline, req.IsArchived); err != nil {
		switch {
		case errors.Is(err, goal.ErrGoalNotFound):
			writeErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Goal not found", "ERR_GOAL_NOT_FOUND")
		case errors.Is(err, goal.ErrGoalNameExists):
			writeErrorJSON(w, http.StatusConflict, "CONFLICT", "Goal name already exists", "ERR_GOAL_EXISTS")
		case errors.Is(err, goal.ErrInvalidName):
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid goal name", "ERR_INVALID_NAME")
		case errors.Is(err, goal.ErrInvalidTarget):
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid goal target", "ERR_INVALID_TARGET")
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update goal", "ERR_INTERNAL_SERVER")
		}
		return
	}
	writeJSON(w, http.StatusOK, SuccessResponse{Message: "Goal updated successfully"})
}
