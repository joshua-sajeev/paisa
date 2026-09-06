package application

import (
	"context"
	"errors"
	"time"

	"github.com/joshu-sajeev/paisa/internal/security"
	"github.com/joshu-sajeev/paisa/internal/session"
)

// ErrInvalidCredentials indicates that authentication failed.
var ErrInvalidCredentials = errors.New("invalid credentials")

// AuthService handles authentication and session management.
type AuthService struct {
	store             session.SessionStore
	pinHash           string
	sessionTTLMinutes int
}

// NewAuthService creates a new authentication service.
func NewAuthService(
	store session.SessionStore,
	pinHash string,
	sessionTTLMinutes int,
) *AuthService {
	return &AuthService{
		store:             store,
		pinHash:           pinHash,
		sessionTTLMinutes: sessionTTLMinutes,
	}
}

// Login verifies the PIN and creates a new authenticated session.
func (s *AuthService) Login(ctx context.Context, pin string) (*session.Session, error) {
	if err := security.VerifyPIN(pin, s.pinHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := security.GenerateSessionToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	ttl := time.Duration(s.sessionTTLMinutes) * time.Minute

	sess := &session.Session{
		ID:        token,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}

	if err := s.store.Create(ctx, sess); err != nil {
		return nil, err
	}

	return sess, nil
}

// Logout deletes the authenticated session.
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.store.Delete(ctx, sessionID)
}
