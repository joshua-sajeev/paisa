package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerTransactionRoutes(r chi.Router, h *handler.TransactionHandler) {
	r.Route("/transactions", func(sub chi.Router) {
		sub.Post("/", h.HandleCreate)
		sub.Get("/", h.HandleList)
		sub.Get("/{id}", h.HandleGetByID)
		sub.Patch("/{id}", h.HandleUpdate)
		sub.Delete("/{id}", h.HandleDelete)
	})

	r.Route("/accounts/{id}/transactions", func(sub chi.Router) {
		sub.Get("/", h.HandleListByAccount)
	})
}
