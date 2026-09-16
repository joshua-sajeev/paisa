package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerGoalRoutes(r chi.Router, h *handler.GoalHandler) {
	r.Route("/goals", func(sub chi.Router) {
		sub.Post("/", h.Create)
		sub.Get("/", h.List)
		sub.Patch("/{id}", h.Patch)
	})
}
