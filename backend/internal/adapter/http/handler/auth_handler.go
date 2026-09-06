// Package handler provides HTTP adapters for the application.
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/joshu-sajeev/paisa/internal/application"
)

const sessionCookieName = "session_id"

// AuthHandler handles authentication-related HTTP endpoints.
type AuthHandler struct {
	authService *application.AuthService
	validate    *validator.Validate
	logger      *slog.Logger
}

// NewAuthHandler creates a new instance of AuthHandler.
func NewAuthHandler(
	authService *application.AuthService,
	logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
		logger:      logger,
	}
}

// LoginRequest represents the expected request body for user authentication.
type LoginRequest struct {
	PIN string `json:"pin" validate:"required"`
}

// LoginResponse represents the response returned upon successful authentication.
type LoginResponse struct {
	Message string `json:"message"`
}

// Login handles user authentication via PIN and sets a secure session cookie.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(r.Context(), "login request received")

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req LoginRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"failed to decode login request body",
			slog.String("error", err.Error()),
		)

		var maxBytesErr *http.MaxBytesError

		switch {
		case errors.Is(err, io.EOF):
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"INVALID_REQUEST",
				"Request body is required",
				"ERR_MISSING_BODY",
			)

		case errors.As(err, &maxBytesErr):
			writeErrorJSON(
				w,
				http.StatusRequestEntityTooLarge,
				"PAYLOAD_TOO_LARGE",
				"Request body is too large",
				"ERR_BODY_TOO_LARGE",
			)

		default:
			writeErrorJSON(
				w,
				http.StatusBadRequest,
				"INVALID_JSON",
				"Request body contains invalid JSON",
				"ERR_INVALID_JSON",
			)
		}

		return
	}

	// Reject multiple JSON values in the request body.
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		h.logger.WarnContext(
			r.Context(),
			"multiple JSON values in login request body",
		)

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"Request body must contain a single JSON object",
			"ERR_INVALID_JSON",
		)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.WarnContext(
			r.Context(),
			"login validation failed",
			slog.String("error", err.Error()),
		)

		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"PIN is required",
			"ERR_INVALID_PIN",
		)
		return
	}

	sess, err := h.authService.Login(r.Context(), req.PIN)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) {
			h.logger.WarnContext(
				r.Context(),
				"invalid credentials attempt",
			)

			writeErrorJSON(
				w,
				http.StatusUnauthorized,
				"UNAUTHORIZED",
				"Invalid credentials",
				"ERR_INVALID_CREDENTIALS",
			)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"failed to login",
			slog.String("error", err.Error()),
		)

		writeErrorJSON(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"Internal server error",
			"ERR_INTERNAL_SERVER",
		)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  sess.ExpiresAt,
	})

	h.logger.InfoContext(r.Context(), "login successful")

	writeJSON(w, http.StatusOK, LoginResponse{
		Message: "login successful",
	})
}

// Logout invalidates the active session and clears the session cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(r.Context(), "logout request received")

	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		if err := h.authService.Logout(r.Context(), cookie.Value); err != nil {
			h.logger.ErrorContext(
				r.Context(),
				"failed to execute logout in service",
				slog.String("error", err.Error()),
			)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	h.logger.InfoContext(r.Context(), "logout successful")

	writeJSON(w, http.StatusOK, LoginResponse{
		Message: "logout successful",
	})
}
