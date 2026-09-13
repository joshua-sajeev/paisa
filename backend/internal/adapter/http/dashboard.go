package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerDashboardRoutes(r chi.Router, h *handler.DashboardHandler) {
	r.Get("/dashboard", h.Get)
}
