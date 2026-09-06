// Package http provides HTTP adapters for the application.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/session"
)

// HandlerRegistry holds all HTTP handlers and dependencies needed by the router.
type HandlerRegistry struct {
	AccountHandler *handler.AccountHandler
	AuthHandler    *handler.AuthHandler
	SessionStore   session.SessionStore
}

// NewRouter creates and configures the application HTTP router.
func NewRouter(h *HandlerRegistry, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	// Create the login rate limiter once for the lifetime of the server.
	loginLimiter := newLoginRateLimiter(1, 5)
	// Global middleware.
	r.Use(
		middleware.RequestID,
		RequestLogger(logger),
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Authentication routes.
	r.Route("/auth", func(r chi.Router) {
		r.With(
			LoginRateLimitMiddleware(loginLimiter, logger),
		).Post("/login", h.AuthHandler.Login)

		r.Post("/logout", h.AuthHandler.Logout)
	})

	// Protected API routes.
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(AuthMiddleware(
			h.SessionStore,
			logger,
		))

		registerAccountRoutes(r, h.AccountHandler)
	})

	return r
}
