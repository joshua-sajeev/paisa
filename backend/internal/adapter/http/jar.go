package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerJarRoutes(r chi.Router, h *handler.JarHandler) {
	r.Route("/jars", func(sub chi.Router) {
		sub.Post("/", h.Create)
		sub.Get("/", h.List)
		sub.Patch("/allocations", h.UpdateAllocations)
		sub.Patch("/{id}", h.Patch)
	})
}
