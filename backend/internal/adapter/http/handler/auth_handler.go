package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/joshu-sajeev/paisa/internal/config"
	"github.com/joshu-sajeev/paisa/internal/security"
	"github.com/joshu-sajeev/paisa/internal/session"
)

type AuthHandler struct {
	cfg    *config.Config
	store  session.SessionStore
	logger *slog.Logger
}

func NewAuthHandler(cfg *config.Config, store session.SessionStore, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{cfg: cfg, store: store, logger: logger}
}

type loginRequest struct {
	PIN string `json:"pin"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(r.Context(), "invalid login request", slog.String("err", err.Error()))
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request", Code: "ERR_BAD_REQUEST"})
		return
	}

	// verify against stored hash
	phc := h.cfg.AppLock.PINHash
	if err := security.VerifyPIN(req.PIN, phc); err != nil {
		h.logger.WarnContext(r.Context(), "pin verify failed", slog.String("err", err.Error()))
		// generic response
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "UNAUTHENTICATED", Message: "invalid credentials", Code: "ERR_UNAUTH"})
		return
	}

	// create session
	token, err := security.GenerateSessionToken()
	if err != nil {
		h.logger.ErrorContext(r.Context(), "session token generation failed", slog.String("err", err.Error()))
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "INTERNAL", Message: "could not create session", Code: "ERR_INTERNAL"})
		return
	}

	ttl := time.Duration(h.cfg.SessionTTLHours) * time.Hour
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	if h.store != nil {
		if err := h.store.CreateSession(r.Context(), token, ttl); err != nil {
			h.logger.ErrorContext(r.Context(), "session store create failed", slog.String("err", err.Error()))
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "INTERNAL", Message: "could not create session", Code: "ERR_INTERNAL"})
			return
		}
	}

	// set cookie
	cookie := &http.Cookie{
		Name:     "app_session",
		Value:    token,
		HttpOnly: true,
		Secure:   h.cfg.Server.Host != "localhost",
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
	}

	// Do not log token in production; only in debug
	h.logger.InfoContext(r.Context(), "session created")

	http.SetCookie(w, cookie)
	writeJSON(w, http.StatusOK, SuccessResponse{Message: "ok"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("app_session")
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusOK, SuccessResponse{Message: "ok"})
		return
	}
	if h.store != nil {
		_ = h.store.DeleteSession(r.Context(), cookie.Value)
	}
	// remove cookie
	cookie = &http.Cookie{
		Name:     "app_session",
		Value:    "",
		HttpOnly: true,
		Secure:   h.cfg.Server.Host != "localhost",
		Path:     "/",
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)
	writeJSON(w, http.StatusOK, SuccessResponse{Message: "ok"})
}
