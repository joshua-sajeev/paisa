package http_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	apphttp "github.com/joshu-sajeev/paisa/internal/adapter/http"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/session"
	"github.com/stretchr/testify/require"
)

type mockRouterAccountService struct {
	listFn func(context.Context) ([]*account.Account, error)
}

func (m *mockRouterAccountService) Create(
	_ context.Context,
	_ string,
) (*account.Account, error) {
	return nil, nil
}

func (m *mockRouterAccountService) List(
	ctx context.Context,
) ([]*account.Account, error) {
	return m.listFn(ctx)
}

func (m *mockRouterAccountService) Update(
	_ context.Context,
	_ uuid.UUID,
	_ *string,
	_ *bool,
) error {
	return nil
}

type mockRouterAuthService struct {
	loginFn  func(context.Context, string) (*session.Session, error)
	logoutFn func(context.Context, string) error
}

func (m *mockRouterAuthService) Login(
	ctx context.Context,
	pin string,
) (*session.Session, error) {
	return m.loginFn(ctx, pin)
}

func (m *mockRouterAuthService) Logout(
	ctx context.Context,
	sessionID string,
) error {
	return m.logoutFn(ctx, sessionID)
}

type mockRouterSessionStore struct {
	getFn func(context.Context, string) (*session.Session, error)
}

func (m *mockRouterSessionStore) Create(
	_ context.Context,
	_ *session.Session,
) error {
	return nil
}

func (m *mockRouterSessionStore) Get(
	ctx context.Context,
	id string,
) (*session.Session, error) {
	return m.getFn(ctx, id)
}

func (m *mockRouterSessionStore) Delete(
	_ context.Context,
	_ string,
) error {
	return nil
}

func newRouterTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestRouter(
	t *testing.T,
	store session.SessionStore,
	demoMode bool,
) http.Handler {
	t.Helper()

	accountService := &mockRouterAccountService{
		listFn: func(
			_ context.Context,
		) ([]*account.Account, error) {
			return []*account.Account{}, nil
		},
	}

	authService := &mockRouterAuthService{
		loginFn: func(
			_ context.Context,
			_ string,
		) (*session.Session, error) {
			now := time.Now().Round(0)

			return &session.Session{
				ID:        "test-session",
				CreatedAt: now,
				ExpiresAt: now.Add(time.Hour),
			}, nil
		},
		logoutFn: func(
			_ context.Context,
			_ string,
		) error {
			return nil
		},
	}

	accountHandler := handler.NewAccountHandler(
		accountService,
		newRouterTestLogger(),
	)

	authHandler := handler.NewAuthHandler(
		authService,
		newRouterTestLogger(),
	)

	registry := &apphttp.HandlerRegistry{
		AccountHandler: accountHandler,
		AuthHandler:    authHandler,
		SessionStore:   store,
		DemoMode:       demoMode,
	}

	return apphttp.NewRouter(
		registry,
		newRouterTestLogger(),
	)
}

func TestNewRouter(t *testing.T) {
	validSession := &session.Session{
		ID:        "valid-session",
		CreatedAt: time.Now().Round(0),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tests := []struct {
		name       string
		method     string
		path       string
		cookie     string
		demoMode   bool
		getFn      func(context.Context, string) (*session.Session, error)
		wantStatus int
	}{
		{
			name:       "health check",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "logout without session",
			method:     http.MethodPost,
			path:       "/auth/logout",
			wantStatus: http.StatusOK,
		},
		{
			name:   "protected route without session",
			method: http.MethodGet,
			path:   "/api/v1/accounts",
			getFn: func(
				_ context.Context,
				_ string,
			) (*session.Session, error) {
				return nil, session.ErrNotFound
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "protected route with valid session",
			method: http.MethodGet,
			path:   "/api/v1/accounts",
			cookie: "valid-session",
			getFn: func(
				_ context.Context,
				id string,
			) (*session.Session, error) {
				require.Equal(t, "valid-session", id)

				return validSession, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "demo mode bypasses authentication",
			method:   http.MethodGet,
			path:     "/api/v1/accounts",
			demoMode: true,
			getFn: func(
				_ context.Context,
				_ string,
			) (*session.Session, error) {
				t.Fatal("session store should not be called in demo mode")
				return nil, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			path:       "/does-not-exist",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockRouterSessionStore{
				getFn: tt.getFn,
			}

			router := newTestRouter(
				t,
				store,
				tt.demoMode,
			)

			req := httptest.NewRequest(
				tt.method,
				tt.path,
				nil,
			)

			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_id",
					Value: tt.cookie,
				})
			}

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(
				t,
				tt.wantStatus,
				rec.Code,
			)
		})
	}
}
