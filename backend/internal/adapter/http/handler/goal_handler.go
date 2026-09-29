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
	Delete(ctx context.Context, id uuid.UUID) error
	AddContribution(ctx context.Context, goalID uuid.UUID, amount int64, occurredAt time.Time) (*goal.Contribution, error)
	ListContributions(ctx context.Context, goalID uuid.UUID) ([]*goal.Contribution, error)
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

type AddContributionRequest struct {
	Amount     int64     `json:"amount" validate:"required,gt=0"`
	OccurredAt time.Time `json:"occurred_at" validate:"required"`
}

type ContributionResponse struct {
	ID         uuid.UUID `json:"id"`
	GoalID     uuid.UUID `json:"goal_id"`
	Amount     int64     `json:"amount"`
	OccurredAt time.Time `json:"occurred_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewContributionResponse(c *goal.Contribution) ContributionResponse {
	return ContributionResponse{
		ID:         c.ID,
		GoalID:     c.GoalID,
		Amount:     c.Amount,
		OccurredAt: c.OccurredAt,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
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

// Delete handles DELETE /goals/{id}.
func (h *GoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid goal ID format", "ERR_INVALID_ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, goal.ErrGoalNotFound):
			writeErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Goal not found", "ERR_GOAL_NOT_FOUND")
		case errors.Is(err, goal.ErrGoalCannotBeDeleted):
			writeErrorJSON(w, http.StatusConflict, "CONFLICT", "Cannot delete in-progress goal; target not achieved", "ERR_GOAL_IN_PROGRESS")
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete goal", "ERR_INTERNAL_SERVER")
		}
		return
	}

	writeJSON(w, http.StatusOK, SuccessResponse{Message: "Goal deleted successfully"})
}

// AddContribution handles POST /goals/{id}/contributions.
func (h *GoalHandler) AddContribution(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid goal ID format", "ERR_INVALID_ID")
		return
	}

	var req AddContributionRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid contribution request", "ERR_INVALID_INPUT")
		return
	}

	c, err := h.service.AddContribution(r.Context(), id, req.Amount, req.OccurredAt)
	if err != nil {
		switch {
		case errors.Is(err, goal.ErrGoalNotFound):
			writeErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Goal not found", "ERR_GOAL_NOT_FOUND")
		case errors.Is(err, goal.ErrInvalidAmount):
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid contribution amount", "ERR_INVALID_AMOUNT")
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to add contribution", "ERR_INTERNAL_SERVER")
		}
		return
	}

	writeJSON(w, http.StatusCreated, NewContributionResponse(c))
}

// ListContributions handles GET /goals/{id}/contributions.
func (h *GoalHandler) ListContributions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid goal ID format", "ERR_INVALID_ID")
		return
	}

	contributions, err := h.service.ListContributions(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, goal.ErrGoalNotFound):
			writeErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Goal not found", "ERR_GOAL_NOT_FOUND")
		default:
			writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list contributions", "ERR_INTERNAL_SERVER")
		}
		return
	}

	response := make([]ContributionResponse, 0, len(contributions))
	for _, c := range contributions {
		response = append(response, NewContributionResponse(c))
	}
	writeJSON(w, http.StatusOK, response)
}
