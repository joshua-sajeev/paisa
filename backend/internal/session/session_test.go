package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryStore(t *testing.T) {
	store := session.NewInMemoryStore()

	assert.NotNil(t, store)
}

func TestInMemoryStore_Create(t *testing.T) {
	tests := []struct {
		name string
		sess *session.Session
	}{
		{
			name: "valid session",
			sess: &session.Session{
				ID:        "test-session-1",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(10 * time.Minute),
			},
		},
		{
			name: "multiple sessions",
			sess: &session.Session{
				ID:        "test-session-2",
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(10 * time.Minute),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := session.NewInMemoryStore()

			err := store.Create(context.Background(), tt.sess)

			require.NoError(t, err)
		})
	}
}

func TestInMemoryStore_Get(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		sessionID     string
		sess          *session.Session
		expectedError error
		expectSession bool
	}{
		{
			name:      "existing session",
			sessionID: "existing-session",
			sess: &session.Session{
				ID:        "existing-session",
				CreatedAt: now,
				ExpiresAt: now.Add(10 * time.Minute),
			},
			expectSession: true,
		},
		{
			name:      "non-existent session",
			sessionID: "non-existent",
			sess: &session.Session{
				ID:        "existing-session",
				CreatedAt: now,
				ExpiresAt: now.Add(10 * time.Minute),
			},
			expectedError: session.ErrNotFound,
		},
		{
			name:      "expired session",
			sessionID: "expired-session",
			sess: &session.Session{
				ID:        "expired-session",
				CreatedAt: now.Add(-20 * time.Minute),
				ExpiresAt: now.Add(-10 * time.Minute),
			},
			expectedError: session.ErrNotFound,
		},
		{
			name:      "session expiring now",
			sessionID: "expiring-session",
			sess: &session.Session{
				ID:        "expiring-session",
				CreatedAt: now,
				ExpiresAt: now,
			},
			expectedError: session.ErrNotFound,
		},
		{
			name:      "session expiring in future",
			sessionID: "future-session",
			sess: &session.Session{
				ID:        "future-session",
				CreatedAt: now,
				ExpiresAt: now.Add(time.Second),
			},
			expectSession: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := session.NewInMemoryStore()

			err := store.Create(context.Background(), tt.sess)
			require.NoError(t, err)

			retrieved, err := store.Get(
				context.Background(),
				tt.sessionID,
			)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, retrieved)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, retrieved)

			assert.Equal(t, tt.sess.ID, retrieved.ID)
			assert.Equal(t, tt.sess.CreatedAt, retrieved.CreatedAt)
			assert.Equal(t, tt.sess.ExpiresAt, retrieved.ExpiresAt)
		})
	}
}

func TestInMemoryStore_CreateMultiple(t *testing.T) {
	ctx := context.Background()
	store := session.NewInMemoryStore()

	for i := 1; i <= 5; i++ {
		sess := &session.Session{
			ID:        sessionID(i),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := store.Create(ctx, sess)
		require.NoError(t, err)
	}

	for i := 1; i <= 5; i++ {
		sess, err := store.Get(ctx, sessionID(i))

		require.NoError(t, err)
		require.NotNil(t, sess)
		assert.Equal(t, sessionID(i), sess.ID)
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		createSession bool
	}{
		{
			name:          "existing session",
			sessionID:     "existing-session",
			createSession: true,
		},
		{
			name:          "non-existent session",
			sessionID:     "non-existent-session",
			createSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := session.NewInMemoryStore()

			if tt.createSession {
				sess := &session.Session{
					ID:        tt.sessionID,
					CreatedAt: time.Now(),
					ExpiresAt: time.Now().Add(10 * time.Minute),
				}

				err := store.Create(ctx, sess)
				require.NoError(t, err)
			}

			err := store.Delete(ctx, tt.sessionID)

			assert.NoError(t, err)

			_, err = store.Get(ctx, tt.sessionID)
			assert.ErrorIs(t, err, session.ErrNotFound)
		})
	}
}

func TestInMemoryStore_DeleteMultiple(t *testing.T) {
	ctx := context.Background()
	store := session.NewInMemoryStore()

	for i := 1; i <= 3; i++ {
		sess := &session.Session{
			ID:        sessionID(i),
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err := store.Create(ctx, sess)
		require.NoError(t, err)
	}

	err := store.Delete(ctx, sessionID(2))
	require.NoError(t, err)

	sess1, err := store.Get(ctx, sessionID(1))
	require.NoError(t, err)
	assert.NotNil(t, sess1)

	sess2, err := store.Get(ctx, sessionID(2))
	assert.ErrorIs(t, err, session.ErrNotFound)
	assert.Nil(t, sess2)

	sess3, err := store.Get(ctx, sessionID(3))
	require.NoError(t, err)
	assert.NotNil(t, sess3)
}

func TestInMemoryStore_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	store := session.NewInMemoryStore()

	const count = 10

	// Create concurrently.
	createDone := make(chan error, count)

	for i := 1; i <= count; i++ {
		go func(id int) {
			sess := &session.Session{
				ID:        sessionID(id),
				CreatedAt: time.Now(),
				ExpiresAt: time.Now().Add(10 * time.Minute),
			}

			createDone <- store.Create(ctx, sess)
		}(i)
	}

	for range count {
		assert.NoError(t, <-createDone)
	}

	// Get concurrently.
	getDone := make(chan error, count)

	for i := 1; i <= count; i++ {
		go func(id int) {
			sess, err := store.Get(ctx, sessionID(id))
			if err != nil {
				getDone <- err
				return
			}

			if sess == nil {
				getDone <- assert.AnError
				return
			}

			getDone <- nil
		}(i)
	}

	for range count {
		assert.NoError(t, <-getDone)
	}

	// Delete concurrently.
	deleteDone := make(chan error, count)

	for i := 1; i <= count; i++ {
		go func(id int) {
			deleteDone <- store.Delete(ctx, sessionID(id))
		}(i)
	}

	for range count {
		assert.NoError(t, <-deleteDone)
	}

	// Verify deletion.
	for i := 1; i <= count; i++ {
		sess, err := store.Get(ctx, sessionID(i))

		assert.ErrorIs(t, err, session.ErrNotFound)
		assert.Nil(t, sess)
	}
}

func TestInMemoryStore_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := session.NewInMemoryStore()

	sess := &session.Session{
		ID:        "test-session",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	err := store.Create(ctx, sess)
	require.NoError(t, err)

	cancel()

	retrieved, err := store.Get(context.Background(), sess.ID)

	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestInMemoryStore_SessionLifecycle(t *testing.T) {
	ctx := context.Background()
	store := session.NewInMemoryStore()

	const id = "lifecycle-test-session"

	now := time.Now()

	sess := &session.Session{
		ID:        id,
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}

	// Create.
	err := store.Create(ctx, sess)
	require.NoError(t, err)

	// Get.
	retrieved, err := store.Get(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, id, retrieved.ID)

	// Delete.
	err = store.Delete(ctx, id)
	require.NoError(t, err)

	// Verify deletion.
	retrieved, err = store.Get(ctx, id)
	assert.ErrorIs(t, err, session.ErrNotFound)
	assert.Nil(t, retrieved)
}

func sessionID(n int) string {
	return map[int]string{
		1: "session-1",
		2: "session-2",
		3: "session-3",
		4: "session-4",
		5: "session-5",
	}[n]
}
