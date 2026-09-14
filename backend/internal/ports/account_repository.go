// Package ports defines the interfaces required by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

// AccountRepository defines the persistence port for accounts.
type AccountRepository interface {
	// Create creates a new account.
	Create(ctx context.Context, a *account.Account) error

	// List gets all accounts.
	List(ctx context.Context) ([]*account.Account, error)

	// FindByID gets an account by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error)

	// Save persists editable account metadata without changing balance.
	Save(ctx context.Context, a *account.Account) error

	// AdjustBalance atomically changes an account balance by delta.
	AdjustBalance(ctx context.Context, id uuid.UUID, delta int64) error
}
