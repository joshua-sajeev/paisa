// Package session provides in-memory session management.
package session

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrNotFound indicates that a session could not be found.
var ErrNotFound = errors.New("session not found")

// Session represents an authenticated user session.
type Session struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore manages user sessions.
type SessionStore interface {
	// Create stores a new user session.
	Create(ctx context.Context, session *Session) error

	// Get retrieves a session by its ID.
	Get(ctx context.Context, id string) (*Session, error)

	// Delete removes a session by its ID.
	Delete(ctx context.Context, id string) error
}

// InMemoryStore stores sessions in memory.
type InMemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewInMemoryStore creates an in-memory session store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		sessions: make(map[string]*Session),
	}
}

func (s *InMemoryStore) Create(_ context.Context, sess *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[sess.ID] = sess
	return nil
}

func (s *InMemoryStore) Get(_ context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[id]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return nil, ErrNotFound
	}

	return sess, nil
}

func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}
