package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
)

// AllocationRepository defines the persistence port for jar allocations.
type AllocationRepository interface {
	// Create persists a new allocation.
	Create(ctx context.Context, alloc *transaction.Allocation) error

	// DeleteByTransactionID removes all allocations for a transaction.
	DeleteByTransactionID(ctx context.Context, transactionID uuid.UUID) error

	// TODO:
	// ListByTransactionID retrieves all allocations for a transaction.
	// ListByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*transaction.Allocation, error)
}
