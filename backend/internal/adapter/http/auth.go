package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerAuthRoutes(r chi.Router, h *handler.AuthHandler) {
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
}
