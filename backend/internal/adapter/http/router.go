// Package http provides HTTP adapters for the application.
package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
)

// HandlerRegistry holds all HTTP handlers needed by the router.
type HandlerRegistry struct {
	AccountHandler *handler.AccountHandler
}

func NewRouter(h *HandlerRegistry, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(
		middleware.RequestID,
		RequestLogger(logger),
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		registerAccountRoutes(r, h.AccountHandler)
	})

	return r
}
