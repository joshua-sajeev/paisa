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

	// List gets all active accounts.
	List(ctx context.Context) ([]*account.Account, error)

	// Update updates the provided account fields.
	// A nil field is left unchanged.
	Update(ctx context.Context, id uuid.UUID, name *string, isArchived *bool) error
}
