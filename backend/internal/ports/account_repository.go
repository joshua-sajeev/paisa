// Package ports defines the interfaces required by the application layer.
package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

// AccountRepository defines the persistence port for accounts.
type AccountRepository interface {
	// Create creates a new account
	Create(ctx context.Context, a *account.Account) error

	// List gets all accounts.
	List(ctx context.Context) ([]*account.Account, error)

	// UpdateName upates an account name
	UpdateName(ctx context.Context, id uuid.UUID, name string) error

	// SetArchived sets archive status for an account
	SetArchived(ctx context.Context, id uuid.UUID, archived bool) error
}
