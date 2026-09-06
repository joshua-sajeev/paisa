package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/joshu-sajeev/paisa/internal/session"
)

const sessionCookieName = "session_id"

// AuthMiddleware protects routes that require an authenticated session.
func AuthMiddleware(
	sessionStore session.SessionStore,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				logger.WarnContext(
					r.Context(),
					"authentication required",
				)

				writeErrorJSON(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"Authentication required",
					"ERR_AUTH_REQUIRED",
				)
				return
			}

			_, err = sessionStore.Get(r.Context(), cookie.Value)
			if err != nil {
				logger.WarnContext(
					r.Context(),
					"invalid or expired session",
				)

				writeErrorJSON(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"Session is invalid or expired",
					"ERR_INVALID_SESSION",
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ErrorResponse represents an HTTP error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}

func writeErrorJSON(w http.ResponseWriter, status int, err string, message string, code string) {
	writeJSON(w, status, ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	})
}
