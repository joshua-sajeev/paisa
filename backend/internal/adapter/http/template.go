package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerTemplateRoutes(r chi.Router, h *handler.TemplateHandler) {
	r.Route("/templates", func(sub chi.Router) {
		sub.Post("/", h.Create)
		sub.Get("/", h.List)
		sub.Get("/{id}", h.Get)
		sub.Patch("/{id}", h.Patch)
		sub.Delete("/{id}", h.Delete)
	})
}
