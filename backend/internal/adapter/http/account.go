package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerAccountRoutes(r chi.Router, h *handler.AccountHandler) {
	r.Route("/accounts", func(sub chi.Router) {
		sub.Post("/", h.Create)
		sub.Get("/", h.List)
		sub.Put("/{id}", h.UpdateName)
		sub.Post("/{id}/archive", h.Archive)
		sub.Post("/{id}/unarchive", h.Unarchive)
	})
}
