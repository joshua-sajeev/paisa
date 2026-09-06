package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/security"
	"github.com/joshu-sajeev/paisa/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionStore struct {
	createFn func(context.Context, *session.Session) error
	getFn    func(context.Context, string) (*session.Session, error)
	deleteFn func(context.Context, string) error
}

func (m *mockSessionStore) Create(
	ctx context.Context,
	sess *session.Session,
) error {
	return m.createFn(ctx, sess)
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
	return m.deleteFn(ctx, id)
}

func newMockSessionStore() *mockSessionStore {
	return &mockSessionStore{
		createFn: func(
			context.Context,
			*session.Session,
		) error {
			return nil
		},
		getFn: func(
			context.Context,
			string,
		) (*session.Session, error) {
			return nil, session.ErrNotFound
		},
		deleteFn: func(
			context.Context,
			string,
		) error {
			return nil
		},
	}
}

func TestNewAuthService(t *testing.T) {
	store := newMockSessionStore()

	service := application.NewAuthService(
		store,
		"test-hash",
		30,
	)

	require.NotNil(t, service)
}

func TestAuthService_Login(t *testing.T) {
	const pin = "481516"

	pinHash, err := security.HashPIN(pin)
	require.NoError(t, err)

	tests := []struct {
		name        string
		pin         string
		createFn    func(context.Context, *session.Session) error
		wantErr     error
		wantSession bool
		wantCreate  bool
		sessionTTL  time.Duration
	}{
		{
			name: "success",
			pin:  pin,
			createFn: func(
				ctx context.Context,
				sess *session.Session,
			) error {
				return nil
			},
			wantSession: true,
			wantCreate:  true,
			sessionTTL:  30 * time.Minute,
		},
		{
			name: "invalid PIN",
			pin:  "123456",
			createFn: func(
				ctx context.Context,
				sess *session.Session,
			) error {
				t.Fatal("Create should not be called")
				return nil
			},
			wantErr:     application.ErrInvalidCredentials,
			wantSession: false,
			wantCreate:  false,
		},
		{
			name: "session store error",
			pin:  pin,
			createFn: func(
				ctx context.Context,
				sess *session.Session,
			) error {
				return errors.New("create session failed")
			},
			wantSession: false,
			wantCreate:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createCalled := false

			store := newMockSessionStore()
			store.createFn = func(
				ctx context.Context,
				sess *session.Session,
			) error {
				createCalled = true

				require.NotNil(t, sess)
				assert.NotEmpty(t, sess.ID)
				assert.False(t, sess.CreatedAt.IsZero())
				assert.False(t, sess.ExpiresAt.IsZero())
				assert.True(t, sess.ExpiresAt.After(sess.CreatedAt))

				return tt.createFn(ctx, sess)
			}

			service := application.NewAuthService(
				store,
				pinHash,
				30,
			)

			before := time.Now()

			sess, err := service.Login(
				context.Background(),
				tt.pin,
			)

			after := time.Now()

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, sess)
				assert.Equal(t, tt.wantCreate, createCalled)
				return
			}

			if tt.name == "session store error" {
				require.Error(t, err)
				assert.Nil(t, sess)
				assert.True(t, createCalled)
				return
			}

			require.NoError(t, err)

			if !tt.wantSession {
				assert.Nil(t, sess)
				return
			}

			require.NotNil(t, sess)

			assert.NotEmpty(t, sess.ID)
			assert.True(
				t,
				!sess.CreatedAt.Before(before) &&
					!sess.CreatedAt.After(after),
			)

			expectedTTL := 30 * time.Minute

			assert.WithinDuration(
				t,
				sess.CreatedAt.Add(expectedTTL),
				sess.ExpiresAt,
				time.Second,
			)

			assert.True(t, createCalled)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		deleteFn  func(context.Context, string) error
		wantErr   bool
	}{
		{
			name:      "success",
			sessionID: "valid-session-token",
			deleteFn: func(
				ctx context.Context,
				sessionID string,
			) error {
				assert.Equal(
					t,
					"valid-session-token",
					sessionID,
				)

				return nil
			},
			wantErr: false,
		},
		{
			name:      "session store error",
			sessionID: "valid-session-token",
			deleteFn: func(
				ctx context.Context,
				sessionID string,
			) error {
				return errors.New("delete session failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteCalled := false

			store := newMockSessionStore()
			store.deleteFn = func(
				ctx context.Context,
				sessionID string,
			) error {
				deleteCalled = true
				return tt.deleteFn(ctx, sessionID)
			}

			service := application.NewAuthService(
				store,
				"unused",
				30,
			)

			err := service.Logout(
				context.Background(),
				tt.sessionID,
			)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.True(t, deleteCalled)
		})
	}
}
