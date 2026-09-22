package http

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/session"
	"github.com/stretchr/testify/require"
)

type mockStaticSessionStore struct {
	getFn func(context.Context, string) (*session.Session, error)
}

func (m *mockStaticSessionStore) Create(
	_ context.Context,
	_ *session.Session,
) error {
	return nil
}

func (m *mockStaticSessionStore) Get(
	ctx context.Context,
	id string,
) (*session.Session, error) {
	return m.getFn(ctx, id)
}

func (m *mockStaticSessionStore) Delete(
	_ context.Context,
	_ string,
) error {
	return nil
}

func TestStaticFrontendHandler(t *testing.T) {
	staticDir := t.TempDir()

	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(staticDir, "index.html"),
			[]byte("index"),
			0o600,
		),
	)
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(staticDir, "login.html"),
			[]byte("login"),
			0o600,
		),
	)
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(staticDir, "dashboard.html"),
			[]byte("dashboard"),
			0o600,
		),
	)
	require.NoError(t, os.MkdirAll(filepath.Join(staticDir, "_next"), 0o700))
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(staticDir, "_next", "app.js"),
			[]byte("asset"),
			0o600,
		),
	)

	validSession := &session.Session{
		ID:        "valid-session",
		CreatedAt: time.Now().Round(0),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tests := []struct {
		name       string
		path       string
		cookie     string
		getFn      func(context.Context, string) (*session.Session, error)
		wantStatus int
		wantBody   string
		wantHeader string
	}{
		{
			name:       "public root serves index",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   "index",
		},
		{
			name:       "login remains public",
			path:       "/login",
			wantStatus: http.StatusOK,
			wantBody:   "login",
		},
		{
			name:       "assets remain public",
			path:       "/_next/app.js",
			wantStatus: http.StatusOK,
			wantBody:   "asset",
		},
		{
			name:       "protected html redirects without session",
			path:       "/dashboard",
			wantStatus: http.StatusFound,
			wantHeader: "/login",
		},
		{
			name:   "protected html serves with valid session",
			path:   "/dashboard",
			cookie: "valid-session",
			getFn: func(
				_ context.Context,
				id string,
			) (*session.Session, error) {
				require.Equal(t, "valid-session", id)
				return validSession, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   "dashboard",
		},
		{
			name:       "path traversal does not escape static dir",
			path:       "/../secret.txt",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStaticSessionStore{
				getFn: func(
					_ context.Context,
					_ string,
				) (*session.Session, error) {
					return nil, session.ErrNotFound
				},
			}
			if tt.getFn != nil {
				store.getFn = tt.getFn
			}

			handler := newStaticFrontendHandler(
				staticDir,
				store,
				false,
				slog.New(slog.NewTextHandler(io.Discard, nil)),
			)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{
					Name:  sessionCookieName,
					Value: tt.cookie,
				})
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != "" {
				require.Equal(t, tt.wantBody, rec.Body.String())
			}
			if tt.wantHeader != "" {
				require.Equal(t, tt.wantHeader, rec.Header().Get("Location"))
			}
		})
	}
}
