package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerStatementRoutes(r chi.Router, h *handler.StatementHandler) {
	r.Get("/accounts/{id}/statement", h.HandleGet)
}
