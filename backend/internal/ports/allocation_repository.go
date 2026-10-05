package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
)

// AllocationListItem represents an allocation joined with its transaction and jar.
type AllocationListItem struct {
	ID              uuid.UUID
	TransactionID   uuid.UUID
	TransactionName string
	JarID           uuid.UUID
	JarName         string
	AllocationType  jar.AllocationType
	Amount          int64
	OccurredAt      time.Time
	CreatedAt       time.Time
}

// AllocationListParams defines filtering and pagination for listing allocations.
type AllocationListParams struct {
	Limit  int
	Offset int

	Search          *string
	JarID           *uuid.UUID
	JarName         *string
	TransactionName *string
	AllocationType  *jar.AllocationType
	FromDate        *time.Time
	ToDate          *time.Time
	MinAmount       *int64
	MaxAmount       *int64
}

// AllocationMonthlySummary contains monthly allocation aggregates.
type AllocationMonthlySummary struct {
	Month             string // YYYY-MM formatted in IST
	TotalAllocated    int64  // Total amount allocated across jars in paise
	JarTotalAllocated *int64 // Total amount allocated to specific jar if filtered
}

// AllocationRepository defines the persistence port for jar allocations.
type AllocationRepository interface {
	// Create persists a new allocation.
	Create(ctx context.Context, alloc *transaction.Allocation) error

	// DeleteByTransactionID removes all allocations for a transaction.
	DeleteByTransactionID(ctx context.Context, transactionID uuid.UUID) error

	// List retrieves allocations with pagination and filtering.
	List(ctx context.Context, params AllocationListParams) ([]*AllocationListItem, int, error)

	// GetMonthlySummary returns total allocated paise for a given date range [start, end).
	// If jarID is non-nil, totals only allocations for that jar.
	GetMonthlySummary(ctx context.Context, start time.Time, end time.Time, jarID *uuid.UUID) (int64, error)

	// SumByJar returns the total amount allocated to a jar within the given date range.
	SumByJar(
		ctx context.Context,
		jarID uuid.UUID,
		startDate *time.Time,
		endDate *time.Time,
	) (int64, error)
}
