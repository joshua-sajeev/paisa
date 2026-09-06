package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apphttp "github.com/joshu-sajeev/paisa/internal/adapter/http"
	"github.com/joshu-sajeev/paisa/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionStore struct {
	getFn func(context.Context, string) (*session.Session, error)
}

func (m *mockSessionStore) Create(
	ctx context.Context,
	sess *session.Session,
) error {
	return nil
}

func (m *mockSessionStore) Get(
	ctx context.Context,
	id string,
) (*session.Session, error) {
	return m.getFn(ctx, id)
}

func (m *mockSessionStore) Delete(
	ctx context.Context,
	id string,
) error {
	return nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		sessionID      string
		getFn          func(context.Context, string) (*session.Session, error)
		wantStatus     int
		wantNextCalled bool
		wantError      string
		wantMessage    string
		wantCode       string
	}{
		{
			name:      "valid session",
			sessionID: "valid-session-token",
			getFn: func(
				ctx context.Context,
				id string,
			) (*session.Session, error) {
				assert.Equal(t, "valid-session-token", id)

				return &session.Session{
					ID:        id,
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(10 * time.Minute),
				}, nil
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:      "invalid session",
			sessionID: "invalid-session-token",
			getFn: func(
				ctx context.Context,
				id string,
			) (*session.Session, error) {
				return nil, session.ErrNotFound
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
			wantError:      "UNAUTHORIZED",
			wantMessage:    "Session is invalid or expired",
			wantCode:       "ERR_INVALID_SESSION",
		},
		{
			name:      "session store error",
			sessionID: "session-token",
			getFn: func(
				ctx context.Context,
				id string,
			) (*session.Session, error) {
				return nil, errors.New("session store unavailable")
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
			wantError:      "UNAUTHORIZED",
			wantMessage:    "Session is invalid or expired",
			wantCode:       "ERR_INVALID_SESSION",
		},
		{
			name:           "missing session cookie",
			sessionID:      "",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
			wantError:      "UNAUTHORIZED",
			wantMessage:    "Authentication required",
			wantCode:       "ERR_AUTH_REQUIRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockSessionStore{
				getFn: tt.getFn,
			}

			nextCalled := false

			next := http.HandlerFunc(func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := apphttp.AuthMiddleware(
				store,
				newTestLogger(),
			)

			handler := middleware(next)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/accounts",
				nil,
			)

			if tt.sessionID != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tt.sessionID,
				})
			}

			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantNextCalled, nextCalled)

			if tt.wantError == "" {
				return
			}

			var response struct {
				Error   string `json:"error"`
				Message string `json:"message"`
				Code    string `json:"code"`
			}

			require.NoError(
				t,
				json.NewDecoder(w.Body).Decode(&response),
			)

			assert.Equal(t, tt.wantError, response.Error)
			assert.Equal(t, tt.wantMessage, response.Message)
			assert.Equal(t, tt.wantCode, response.Code)
		})
	}
}
