package security

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting login attempts.
type RateLimiter interface {
	// AllowLogin checks if a login attempt is allowed for the given identifier.
	// Returns true if allowed, false if rate limited.
	AllowLogin(ctx context.Context, identifier string) (bool, error)
	
	// RecordFailure records a failed login attempt for the identifier.
	RecordFailure(ctx context.Context, identifier string) error
	
	// ResetFailures resets the failure count for the identifier (after successful login).
	ResetFailures(ctx context.Context, identifier string) error
}

// InMemoryRateLimiter is a simple in-memory rate limiter for login attempts.
// It uses exponential backoff: 1s, 2s, 4s, 8s, 16s after each failure.
// After 5 failed attempts, the account is locked for 15 minutes.
type InMemoryRateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*attemptRecord
}

type attemptRecord struct {
	failureCount int
	lastFailure  time.Time
	lockedUntil  time.Time
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter.
func NewInMemoryRateLimiter() *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		attempts: make(map[string]*attemptRecord),
	}
}

// AllowLogin checks if a login attempt is allowed.
func (rl *InMemoryRateLimiter) AllowLogin(_ context.Context, identifier string) (bool, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	record, exists := rl.attempts[identifier]
	if !exists {
		return true, nil
	}

	// Check if locked
	if record.lockedUntil.After(time.Now()) {
		return false, fmt.Errorf("account locked: try again in %v", time.Until(record.lockedUntil).Round(time.Second))
	}

	// Check exponential backoff
	backoffDuration := rl.calculateBackoff(record.failureCount)
	if time.Since(record.lastFailure) < backoffDuration {
		return false, fmt.Errorf("too many attempts: try again in %v", backoffDuration-time.Since(record.lastFailure).Round(time.Second))
	}

	return true, nil
}

// RecordFailure records a failed login attempt.
func (rl *InMemoryRateLimiter) RecordFailure(_ context.Context, identifier string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	record, exists := rl.attempts[identifier]
	if !exists {
		record = &attemptRecord{}
		rl.attempts[identifier] = record
	}

	record.failureCount++
	record.lastFailure = time.Now()

	// Lock account after 5 failures for 15 minutes
	if record.failureCount >= 5 {
		record.lockedUntil = time.Now().Add(15 * time.Minute)
	}

	return nil
}

// ResetFailures resets the failure count for the identifier (successful login).
func (rl *InMemoryRateLimiter) ResetFailures(_ context.Context, identifier string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	delete(rl.attempts, identifier)
	return nil
}

// calculateBackoff returns exponential backoff duration: 1s, 2s, 4s, 8s, 16s for attempts 1-5+.
func (rl *InMemoryRateLimiter) calculateBackoff(failureCount int) time.Duration {
	if failureCount <= 0 {
		return 0
	}
	if failureCount > 5 {
		failureCount = 5 // Cap at 5 to avoid huge delays
	}

	// 1 << (failureCount - 1) = 2^(failureCount-1)
	// failure 1: 2^0 = 1s
	// failure 2: 2^1 = 2s
	// failure 3: 2^2 = 4s
	// failure 4: 2^3 = 8s
	// failure 5: 2^4 = 16s
	return time.Duration(1<<(failureCount-1)) * time.Second
}
