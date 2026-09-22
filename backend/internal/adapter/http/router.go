// Package http provides HTTP adapters for the application.
package http

import (
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	loginLimiter := newLoginRateLimiter(5, 3)
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

	r.Handle("/*", newStaticFrontendHandler(
		staticDir,
		h.SessionStore,
		h.DemoMode,
		logger,
	))
	return r
}

var protectedHTMLRoutes = map[string]struct{}{
	"/dashboard":    {},
	"/transactions": {},
	"/accounts":     {},
	"/jars":         {},
	"/goals":        {},
}

func newStaticFrontendHandler(
	staticDir string,
	sessionStore session.SessionStore,
	demoMode bool,
	logger *slog.Logger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		cleanPath := cleanURLPath(req.URL.Path)

		if _, ok := protectedHTMLRoutes[cleanPath]; ok &&
			!hasValidSession(req, sessionStore, demoMode, logger) {
			http.Redirect(w, req, "/login", http.StatusFound)
			return
		}

		file, ok := staticFilePath(staticDir, htmlPath(cleanPath))
		if ok {
			if _, err := os.Stat(file); err == nil {
				http.ServeFile(w, req, file)
				return
			}
		}

		asset, ok := staticFilePath(staticDir, cleanPath)
		if ok {
			if _, err := os.Stat(asset); err == nil {
				http.ServeFile(w, req, asset)
				return
			}
		}

		http.NotFound(w, req)
	})
}

func hasValidSession(
	req *http.Request,
	sessionStore session.SessionStore,
	demoMode bool,
	logger *slog.Logger,
) bool {
	if demoMode {
		return true
	}

	cookie, err := req.Cookie(sessionCookieName)
	if err != nil {
		logger.WarnContext(req.Context(), "authentication required")
		return false
	}

	if _, err := sessionStore.Get(req.Context(), cookie.Value); err != nil {
		logger.WarnContext(req.Context(), "invalid or expired session")
		return false
	}

	return true
}

func htmlPath(requestPath string) string {
	if requestPath == "/" {
		return "/index.html"
	}

	return requestPath + ".html"
}

func cleanURLPath(requestPath string) string {
	return path.Clean("/" + requestPath)
}

func staticFilePath(staticDir string, requestPath string) (string, bool) {
	relativePath := strings.TrimPrefix(cleanURLPath(requestPath), "/")
	if relativePath == "." || strings.HasPrefix(relativePath, "../") {
		return "", false
	}

	return filepath.Join(staticDir, filepath.FromSlash(relativePath)), true
}
