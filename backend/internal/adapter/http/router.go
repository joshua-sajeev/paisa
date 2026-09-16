// Package http provides HTTP adapters for the application.
package http

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/session"
)

// HandlerRegistry holds all HTTP handlers and dependencies needed by the router.
type HandlerRegistry struct {
	DashboardHandler   *handler.DashboardHandler
	AccountHandler     *handler.AccountHandler
	JarHandler         *handler.JarHandler
	TransactionHandler *handler.TransactionHandler
	AuthHandler        *handler.AuthHandler
	GoalHandler        *handler.GoalHandler
	SessionStore       session.SessionStore
	SessionHandler     *handler.SessionHandler
	DemoMode           bool
}

// NewRouter creates and configures the application HTTP router.
func NewRouter(h *HandlerRegistry, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	// Create the login rate limiter once for the lifetime of the server.
	loginLimiter := newLoginRateLimiter(5, 1)
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
		registerAuthRoutes(r, h.AuthHandler, loginLimiter, logger)
	})

	// Protected API routes.
	r.Route("/api/v1", func(r chi.Router) {
		// Public session endpoint.
		registerSessionRoutes(r, h.SessionHandler, h.SessionStore)

		// Protected routes.
		r.Group(func(r chi.Router) {
			if !h.DemoMode {
				r.Use(AuthMiddleware(
					h.SessionStore,
					logger,
				))
			}

			registerAccountRoutes(r, h.AccountHandler)
			registerJarRoutes(r, h.JarHandler)
			registerGoalRoutes(r, h.GoalHandler)
			registerTransactionRoutes(r, h.TransactionHandler)
			registerDashboardRoutes(r, h.DashboardHandler)
		})
	})
	// Static frontend.
	staticDir := "../frontend/out"

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		if path == "/" {
			path = "/index.html"
		} else {
			path += ".html"
		}

		file := filepath.Join(staticDir, filepath.Clean(path))

		if _, err := os.Stat(file); err == nil {
			http.ServeFile(w, req, file)
			return
		}

		asset := filepath.Join(staticDir, filepath.Clean(req.URL.Path))

		if _, err := os.Stat(asset); err == nil {
			http.ServeFile(w, req, asset)
			return
		}

		http.NotFound(w, req)
	}))
	return r
}
