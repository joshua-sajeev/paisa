package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/session"
)

func registerSessionRoutes(
	r chi.Router,
	h *handler.SessionHandler,
	sessionStore session.SessionStore,
) {
	r.Get("/session", h.GetSession(sessionStore))
}
