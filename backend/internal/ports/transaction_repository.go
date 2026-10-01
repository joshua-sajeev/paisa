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
	// JarID restricts results to transactions assigned to a single jar.
	JarID *uuid.UUID
	// FromDate is an inclusive lower bound on occurred_at.
	FromDate *time.Time
	// ToDate is an exclusive upper bound on occurred_at.
	ToDate    *time.Time
	MinAmount *int64
	MaxAmount *int64
}

// TransactionListItem represents a transaction projected for list responses.
type TransactionListItem struct {
	ID             uuid.UUID
	Name           string
	Type           transaction.TransactionType
	Amount         int64
	JarName        *string
	Account        string
	AccountBalance *int64
	Category       transaction.TransactionCategory
	OccurredAt     time.Time
}

// StatementSummary holds balance aggregates for one account over a date range.
// All values are in paise. Outflow is reported as a positive number.
//
// OpeningBalance + Inflow - Outflow == ClosingBalance always holds.
type StatementSummary struct {
	// OpeningBalance is the account balance immediately before the range starts.
	OpeningBalance int64
	// ClosingBalance is the account balance at the end of the range.
	ClosingBalance int64
	// Inflow is the total money received by the account within the range.
	Inflow int64
	// Outflow is the total money leaving the account within the range.
	Outflow int64
}

// Net returns the net change in balance over the range.
func (s StatementSummary) Net() int64 {
	return s.Inflow - s.Outflow
}

// TransactionRepository defines the persistence port for transactions.
type TransactionRepository interface {
	// Create persists a new transaction.
	Create(ctx context.Context, t *transaction.Transaction) error

	// List retrieves transactions with optional filtering and pagination.
	// Returns display-ready transaction list items.
	// Supports filtering by search, type, category, date range, amount range, etc.
	// Results are ordered newest-first by occurred_at, created_at, id.
	List(ctx context.Context, params ListParams) ([]*TransactionListItem, error)

	// ListByAccount retrieves transactions for a specific account with running balance.
	// The running balance is relative to the account: positive for money in, negative for money out.
	// Filters by from_account_id OR to_account_id automatically.
	// Results are ordered newest-first by occurred_at, created_at, id.
	// Balance correctness is maintained across pagination.
	ListByAccount(
		ctx context.Context,
		accountID uuid.UUID,
		params ListParams,
	) ([]*TransactionListItem, error)

	// CountByAccount returns how many transactions for the account match the
	// filters in params. Limit and Offset are ignored.
	CountByAccount(
		ctx context.Context,
		accountID uuid.UUID,
		params ListParams,
	) (int, error)

	// SummarizeByAccount returns balance aggregates for the account within
	// [from, to). A nil bound is open-ended. Only the date range applies;
	// search/type/category/jar filters are intentionally not part of this.
	SummarizeByAccount(
		ctx context.Context,
		accountID uuid.UUID,
		from *time.Time,
		to *time.Time,
	) (StatementSummary, error)

	// FindByID retrieves a single transaction by ID.
	FindByID(ctx context.Context, id uuid.UUID) (*transaction.Transaction, error)

	// Save persists changes to an existing transaction.
	Save(ctx context.Context, t *transaction.Transaction) error

	// Delete removes a transaction by ID.
	Delete(ctx context.Context, id uuid.UUID) error
}
