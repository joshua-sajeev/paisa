package handler

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/seed"
)

type DemoHandler struct {
	db      *pgxpool.Pool
	enabled bool
	logger  *slog.Logger
	resetMu sync.Mutex
}

func NewDemoHandler(db *pgxpool.Pool, enabled bool, logger *slog.Logger) *DemoHandler {
	return &DemoHandler{db: db, enabled: enabled, logger: logger}
}

func (h *DemoHandler) Mode(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, struct {
		Enabled bool `json:"enabled"`
	}{Enabled: h.enabled})
}

func (h *DemoHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if !h.enabled {
		writeErrorJSON(w, http.StatusNotFound, "NOT_FOUND", "Demo mode is disabled", "DEMO_MODE_DISABLED")
		return
	}

	h.resetMu.Lock()
	defer h.resetMu.Unlock()

	if err := seed.Run(r.Context(), h.db); err != nil {
		h.logger.ErrorContext(r.Context(), "failed to reset demo data", slog.String("error", err.Error()))
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to reset demo data", "DEMO_RESET_FAILED")
		return
	}

	h.logger.InfoContext(r.Context(), "demo data reset completed")
	writeJSON(w, http.StatusOK, SuccessResponse{Message: "Demo data reset successfully"})
}
