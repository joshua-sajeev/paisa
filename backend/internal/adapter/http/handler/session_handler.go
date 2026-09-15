// Package handler provides HTTP adapters for the application.
package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/joshu-sajeev/paisa/internal/session"
)

// SessionStore defines the session operations required by SessionHandler.
type SessionStore interface {
	Get(context.Context, string) (*session.Session, error)
}

// SessionHandler handles session-related HTTP endpoints.
type SessionHandler struct {
	demoMode bool
	logger   *slog.Logger
}

// NewSessionHandler creates a new instance of SessionHandler.
func NewSessionHandler(demoMode bool, logger *slog.Logger) *SessionHandler {
	return &SessionHandler{
		demoMode: demoMode,
		logger:   logger,
	}
}

// SessionResponse represents the response for session status check.
type SessionResponse struct {
	Authenticated bool `json:"authenticated"`
}

// GetSession checks if the user has a valid session and returns the status.
// It reads the session cookie and validates it against the session store.
// If demo mode is enabled, it returns authenticated=true without requiring a session.
func (h *SessionHandler) GetSession(
	sessionStore SessionStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In demo mode, always return authenticated
		if h.demoMode {
			h.logger.DebugContext(r.Context(), "demo mode enabled, returning authenticated=true")
			writeJSON(w, http.StatusOK, SessionResponse{
				Authenticated: true,
			})
			return
		}

		// Try to get the session cookie
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			h.logger.DebugContext(
				r.Context(),
				"no session cookie found",
				slog.String("error", err.Error()),
			)
			writeJSON(w, http.StatusOK, SessionResponse{
				Authenticated: false,
			})
			return
		}

		// Validate the session in the store
		_, err = sessionStore.Get(r.Context(), cookie.Value)
		if err != nil {
			h.logger.DebugContext(
				r.Context(),
				"invalid or expired session",
				slog.String("error", err.Error()),
			)
			writeJSON(w, http.StatusOK, SessionResponse{
				Authenticated: false,
			})
			return
		}

		h.logger.DebugContext(r.Context(), "valid session found")
		writeJSON(w, http.StatusOK, SessionResponse{
			Authenticated: true,
		})
	}
}
