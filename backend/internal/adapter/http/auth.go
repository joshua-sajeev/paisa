package http

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

func registerAuthRoutes(
	r chi.Router,
	h *handler.AuthHandler,
	loginLimiter *loginRateLimiter,
	logger *slog.Logger,
) {
	r.With(
		LoginRateLimitMiddleware(loginLimiter, logger),
	).Post("/login", h.Login)

	r.Post("/logout", h.Logout)
}
