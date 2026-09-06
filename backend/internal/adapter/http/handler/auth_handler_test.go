package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	loginFn  func(context.Context, string) (*session.Session, error)
	logoutFn func(context.Context, string) error
}

func (m *mockAuthService) Login(
	ctx context.Context,
	pin string,
) (*session.Session, error) {
	return m.loginFn(ctx, pin)
}

func (m *mockAuthService) Logout(
	ctx context.Context,
	sessionID string,
) error {
	return m.logoutFn(ctx, sessionID)
}

func newTestAuthHandler(svc *mockAuthService) *handler.AuthHandler {
	logger := slog.New(slog.DiscardHandler)
	return handler.NewAuthHandler(svc, logger)
}

func newTestSession() *session.Session {
	now := time.Now()

	return &session.Session{
		ID:        "valid-session-token-123",
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
	}
}

func TestNewAuthHandler(t *testing.T) {
	mockSvc := &mockAuthService{
		loginFn: func(
			ctx context.Context,
			pin string,
		) (*session.Session, error) {
			return newTestSession(), nil
		},
		logoutFn: func(
			ctx context.Context,
			sessionID string,
		) error {
			return nil
		},
	}

	logger := slog.New(slog.DiscardHandler)

	authHandler := handler.NewAuthHandler(mockSvc, logger)

	assert.NotNil(t, authHandler)
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		loginFn         func(context.Context, string) (*session.Session, error)
		wantStatus      int
		wantMessage     string
		wantCookie      bool
		wantCookieValue string
	}{
		{
			name: "success",
			body: `{"pin":"654321"}`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				assert.Equal(t, "654321", pin)
				return newTestSession(), nil
			},
			wantStatus:      http.StatusOK,
			wantMessage:     "login successful",
			wantCookie:      true,
			wantCookieValue: "valid-session-token-123",
		},
		{
			name: "invalid credentials",
			body: `{"pin":"654321"}`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				return nil, application.ErrInvalidCredentials
			},
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "Invalid credentials",
		},
		{
			name: "service error",
			body: `{"pin":"654321"}`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				return nil, errors.New("service error")
			},
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "Internal server error",
		},
		{
			name: "malformed JSON",
			body: `{"pin":`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				t.Fatal("Login should not be called")
				return nil, nil
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Request body contains invalid JSON",
		},
		{
			name: "missing PIN",
			body: `{"pin":""}`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				t.Fatal("Login should not be called")
				return nil, nil
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "PIN is required",
		},
		{
			name: "missing body",
			body: ``,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				t.Fatal("Login should not be called")
				return nil, nil
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Request body is required",
		},
		{
			name: "unknown field",
			body: `{"pin":"654321","foo":"bar"}`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				t.Fatal("Login should not be called")
				return nil, nil
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Request body contains invalid JSON",
		},
		{
			name: "multiple JSON values",
			body: `{"pin":"654321"}{"pin":"654321"}`,
			loginFn: func(
				ctx context.Context,
				pin string,
			) (*session.Session, error) {
				t.Fatal("Login should not be called")
				return nil, nil
			},
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Request body must contain a single JSON object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAuthService{
				loginFn: tt.loginFn,
				logoutFn: func(
					ctx context.Context,
					sessionID string,
				) error {
					return nil
				},
			}

			authHandler := newTestAuthHandler(mockSvc)

			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/login",
				bytes.NewBufferString(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			authHandler.Login(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response struct {
				Error   string `json:"error"`
				Message string `json:"message"`
				Code    string `json:"code"`
			}

			require.NoError(
				t,
				json.NewDecoder(w.Body).Decode(&response),
			)

			assert.Equal(t, tt.wantMessage, response.Message)

			cookies := w.Result().Cookies()

			if !tt.wantCookie {
				assert.Empty(t, cookies)
				return
			}

			require.Len(t, cookies, 1)

			cookie := cookies[0]

			assert.Equal(t, "session_id", cookie.Name)
			assert.Equal(t, tt.wantCookieValue, cookie.Value)
			assert.Equal(t, "/", cookie.Path)
			assert.True(t, cookie.HttpOnly)
			assert.True(t, cookie.Secure)
			assert.Equal(
				t,
				http.SameSiteStrictMode,
				cookie.SameSite,
			)
			assert.False(t, cookie.Expires.IsZero())
		})
	}
}

func TestAuthHandler_Login_BodyTooLarge(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{
			name: "body exceeds maximum size",
			body: []byte(
				`{"pin":"` +
					strings.Repeat("1", (1<<20)+1) +
					`"}`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAuthService{
				loginFn: func(
					ctx context.Context,
					pin string,
				) (*session.Session, error) {
					t.Fatal("Login should not be called")
					return nil, nil
				},
				logoutFn: func(
					ctx context.Context,
					sessionID string,
				) error {
					return nil
				},
			}

			authHandler := newTestAuthHandler(mockSvc)

			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/login",
				bytes.NewReader(tt.body),
			)

			w := httptest.NewRecorder()

			authHandler.Login(w, req)

			assert.Equal(
				t,
				http.StatusRequestEntityTooLarge,
				w.Code,
			)

			var response struct {
				Error   string `json:"error"`
				Message string `json:"message"`
				Code    string `json:"code"`
			}

			require.NoError(
				t,
				json.NewDecoder(w.Body).Decode(&response),
			)

			assert.Equal(
				t,
				"PAYLOAD_TOO_LARGE",
				response.Error,
			)

			assert.Equal(
				t,
				"Request body is too large",
				response.Message,
			)

			assert.Equal(
				t,
				"ERR_BODY_TOO_LARGE",
				response.Code,
			)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		logoutFn       func(context.Context, string) error
		wantStatus     int
		wantMessage    string
		wantLogoutCall bool
	}{
		{
			name:      "success",
			sessionID: "valid-session-token",
			logoutFn: func(
				ctx context.Context,
				sessionID string,
			) error {
				assert.Equal(t, "valid-session-token", sessionID)
				return nil
			},
			wantStatus:     http.StatusOK,
			wantMessage:    "logout successful",
			wantLogoutCall: true,
		},
		{
			name:        "no session cookie",
			sessionID:   "",
			wantStatus:  http.StatusOK,
			wantMessage: "logout successful",
		},
		{
			name:      "service error",
			sessionID: "valid-session-token",
			logoutFn: func(
				ctx context.Context,
				sessionID string,
			) error {
				return errors.New("logout failed")
			},
			wantStatus:     http.StatusOK,
			wantMessage:    "logout successful",
			wantLogoutCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logoutCalled := false

			mockSvc := &mockAuthService{
				loginFn: func(
					ctx context.Context,
					pin string,
				) (*session.Session, error) {
					return newTestSession(), nil
				},
				logoutFn: func(
					ctx context.Context,
					sessionID string,
				) error {
					logoutCalled = true

					if tt.logoutFn != nil {
						return tt.logoutFn(ctx, sessionID)
					}

					return nil
				},
			}

			authHandler := newTestAuthHandler(mockSvc)

			req := httptest.NewRequest(
				http.MethodPost,
				"/auth/logout",
				nil,
			)

			if tt.sessionID != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tt.sessionID,
				})
			}

			w := httptest.NewRecorder()

			authHandler.Logout(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var response struct {
				Message string `json:"message"`
			}

			require.NoError(
				t,
				json.NewDecoder(w.Body).Decode(&response),
			)

			assert.Equal(t, tt.wantMessage, response.Message)
			assert.Equal(t, tt.wantLogoutCall, logoutCalled)

			cookies := w.Result().Cookies()

			require.Len(t, cookies, 1)

			cookie := cookies[0]

			assert.Equal(t, "session_id", cookie.Name)
			assert.Equal(t, "", cookie.Value)
			assert.Equal(t, "/", cookie.Path)
			assert.Equal(t, -1, cookie.MaxAge)
			assert.True(t, cookie.HttpOnly)
			assert.True(t, cookie.Secure)
			assert.Equal(
				t,
				http.SameSiteStrictMode,
				cookie.SameSite,
			)
		})
	}
}
