package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// Statement is the read model behind the account statement screen.
type Statement struct {
	Account *account.Account

	// Summary covers the date range only. It is not affected by the
	// search/type/category/jar filters, so that
	// opening + inflow - outflow == closing always holds.
	Summary ports.StatementSummary

	// Transactions are the filtered, paginated rows, newest first. Amounts are
	// signed relative to the account and each row carries the account balance
	// after that transaction.
	Transactions []*ports.TransactionListItem

	// Total is the number of transactions matching the filters, ignoring
	// pagination.
	Total int
}

// StatementService builds account statements.
type StatementService struct {
	accountRepo     ports.AccountRepository
	transactionRepo ports.TransactionRepository
	logger          *slog.Logger
}

// NewStatementService creates a new StatementService.
func NewStatementService(
	accountRepo ports.AccountRepository,
	transactionRepo ports.TransactionRepository,
	logger *slog.Logger,
) *StatementService {
	return &StatementService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

// Get returns the statement for an account. params.AccountID is ignored; the
// accountID argument is authoritative.
func (s *StatementService) Get(
	ctx context.Context,
	accountID uuid.UUID,
	params ports.ListParams,
) (*Statement, error) {
	acc, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get statement: %w", err)
	}

	summary, err := s.transactionRepo.SummarizeByAccount(
		ctx,
		accountID,
		params.FromDate,
		params.ToDate,
	)
	if err != nil {
		return nil, fmt.Errorf("get statement: %w", err)
	}

	total, err := s.transactionRepo.CountByAccount(ctx, accountID, params)
	if err != nil {
		return nil, fmt.Errorf("get statement: %w", err)
	}

	txns, err := s.transactionRepo.ListByAccount(ctx, accountID, params)
	if err != nil {
		return nil, fmt.Errorf("get statement: %w", err)
	}

	return &Statement{
		Account:      acc,
		Summary:      summary,
		Transactions: txns,
		Total:        total,
	}, nil
}
