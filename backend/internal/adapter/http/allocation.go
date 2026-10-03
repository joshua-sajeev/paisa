package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerAllocationRoutes(r chi.Router, h *handler.AllocationHandler) {
	r.Route("/allocations", func(sub chi.Router) {
		sub.Get("/", h.HandleList)
	})
}
