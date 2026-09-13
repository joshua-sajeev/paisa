package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/joshu-sajeev/paisa/internal/application"
)

type DashboardService interface {
	GetDashboard(ctx context.Context) (*application.DashboardResponse, error)
}

type DashboardHandler struct {
	service DashboardService
	logger  *slog.Logger
}

func NewDashboardHandler(service DashboardService, logger *slog.Logger) *DashboardHandler {
	return &DashboardHandler{
		service: service,
		logger:  logger,
	}
}

// Get handles GET /api/v1/dashboard
func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.DebugContext(ctx, "fetching dashboard")

	dashboard, err := h.service.GetDashboard(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to fetch dashboard", slog.String("error", err.Error()))
		writeErrorJSON(
			w,
			http.StatusInternalServerError,
			"Internal Server Error",
			"Failed to fetch dashboard",
			"DASHBOARD_FETCH_FAILED",
		)
		return
	}

	response := NewDashboardResponse(dashboard)
	h.logger.InfoContext(ctx, "dashboard fetched successfully")
	writeJSON(w, http.StatusOK, response)
}

var _ DashboardService = (*application.DashboardService)(nil)
