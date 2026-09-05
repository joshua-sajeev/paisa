package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/config"
	"github.com/joshu-sajeev/paisa/internal/session"
)

type contextKey string

const AuthenticatedContextKey contextKey = "authenticated"

func AuthMiddleware(
	cfg *config.Config,
	store session.SessionStore,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	if cfg == nil {
		panic("auth middleware: nil config")
	}

	if logger == nil {
		panic("auth middleware: nil logger")
	}

	if !cfg.AppLock.DemoMode && store == nil {
		panic("auth middleware: nil session store")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Demo mode bypasses session authentication.
			if cfg.AppLock.DemoMode {
				ctx := context.WithValue(
					r.Context(),
					AuthenticatedContextKey,
					true,
				)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			cookie, err := r.Cookie("app_session")
			if err != nil || cookie.Value == "" {
				writeJSON(
					w,
					http.StatusUnauthorized,
					handler.ErrorResponse{
						Error:   "UNAUTHENTICATED",
						Message: "authentication required",
						Code:    "ERR_UNAUTH",
					},
				)
				return
			}

			ok, err := store.GetSession(r.Context(), cookie.Value)
			if err != nil {
				logger.ErrorContext(
					r.Context(),
					"session store error",
					slog.String("err", err.Error()),
				)

				writeJSON(
					w,
					http.StatusInternalServerError,
					handler.ErrorResponse{
						Error:   "INTERNAL",
						Message: "internal error",
						Code:    "ERR_INTERNAL",
					},
				)
				return
			}

			if !ok {
				writeJSON(
					w,
					http.StatusUnauthorized,
					handler.ErrorResponse{
						Error:   "UNAUTHENTICATED",
						Message: "invalid session",
						Code:    "ERR_UNAUTH",
					},
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				AuthenticatedContextKey,
				true,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
