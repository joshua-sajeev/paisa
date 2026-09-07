package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
)

// ListParams defines filtering and pagination parameters for transaction queries.
// All filter fields are optional (pointers to zero values mean "no filter").
type ListParams struct {
	// Pagination
	Limit  int // Rows per page; required
	Offset int // Number of rows to skip; required

	Search    *string
	AccountID *uuid.UUID
	Type      *transaction.TransactionType
	Category  *transaction.TransactionCategory
	FromDate  *time.Time
	ToDate    *time.Time
	MinAmount *int64
	MaxAmount *int64
}

// TransactionWithBalance represents a transaction with its running balance.
// Used for account statements where the balance is meaningful.
type TransactionWithBalance struct {
	Transaction  *transaction.Transaction
	BalanceAfter int64 // Running balance after this transaction for the specific account
}

// TransactionRepository defines the persistence port for transactions.
type TransactionRepository interface {
	// Create persists a new transaction.
	Create(ctx context.Context, t *transaction.Transaction) error

	// List retrieves transactions with optional filtering and pagination.
	// Returns plain transactions without running balance (balance is meaningless across multiple accounts).
	// Supports filtering by search, type, category, date range, amount range, etc.
	// Results are ordered newest-first by occurred_at, created_at, id.
	List(ctx context.Context, params ListParams) ([]*transaction.Transaction, error)

	// ListByAccount retrieves transactions for a specific account with running balance.
	// The running balance is relative to the account: positive for money in, negative for money out.
	// Filters by from_account_id OR to_account_id automatically.
	// Results are ordered newest-first by occurred_at, created_at, id.
	// Balance correctness is maintained across pagination.
	ListByAccount(
		ctx context.Context,
		accountID uuid.UUID,
		params ListParams,
	) ([]*TransactionWithBalance, error)

	// FindByID retrieves a single transaction by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*transaction.Transaction, error)

	// Save persists changes to an existing transaction.
	Save(ctx context.Context, t *transaction.Transaction) error

	// Delete removes a transaction by ID.
	Delete(ctx context.Context, id uuid.UUID) error
}
