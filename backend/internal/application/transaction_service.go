package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// TransactionService handles transaction use cases.
type TransactionService struct {
	transactionRepo ports.TransactionRepository
	allocationRepo  ports.AllocationRepository
	jarRepo         ports.JarRepository
	accountRepo     ports.AccountRepository
	txManager       ports.TxManager
	logger          *slog.Logger
}

// NewTransactionService creates a new TransactionService.
func NewTransactionService(
	transactionRepo ports.TransactionRepository,
	allocationRepo ports.AllocationRepository,
	jarRepo ports.JarRepository,
	accountRepo ports.AccountRepository,
	txManager ports.TxManager,
	logger *slog.Logger,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		allocationRepo:  allocationRepo,
		jarRepo:         jarRepo,
		accountRepo:     accountRepo,
		txManager:       txManager,
		logger:          logger,
	}
}

// Create persists a new transaction and allocations (if applicable).
func (s *TransactionService) Create(
	ctx context.Context,
	name string,
	transactionType transaction.TransactionType,
	category transaction.TransactionCategory,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	jarID *uuid.UUID,
	amount int64,
	occurredAt time.Time,
	isMasterIncome bool,
) (*transaction.Transaction, error) {
	s.logger.DebugContext(
		ctx,
		"creating new transaction",
		slog.String("name", name),
		slog.String("type", string(transactionType)),
	)
	if transactionType != transaction.TransactionTypeIncome {
		isMasterIncome = false
	}
	newTxn, err := transaction.NewTransaction(
		name,
		transactionType,
		category,
		fromAccountID,
		toAccountID,
		jarID,
		amount,
		occurredAt,
		isMasterIncome,
	)
	if err != nil {
		s.logger.WarnContext(
			ctx,
			"invalid transaction",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.transactionRepo.Create(txCtx, newTxn); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to create transaction",
				slog.String("error", err.Error()),
			)
			return err
		}

		if isMasterIncome {
			if err := s.createAllocations(txCtx, newTxn); err != nil {
				s.logger.ErrorContext(
					txCtx,
					"failed to create allocations",
					slog.String("error", err.Error()),
				)
				return err
			}
		} else if newTxn.Type == transaction.TransactionTypeIncome && jarID != nil {
			alloc := &transaction.Allocation{
				ID:            uuid.New(),
				TransactionID: newTxn.ID,
				JarID:         *jarID,
				Amount:        amount,
			}

			if err := s.allocationRepo.Create(txCtx, alloc); err != nil {
				s.logger.ErrorContext(
					txCtx,
					"failed to create allocation",
					slog.String("error", err.Error()),
				)
				return err
			}
		}

		if err := s.updateAccountBalances(txCtx, newTxn); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to update account balances",
				slog.String("error", err.Error()),
			)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.InfoContext(
		ctx,
		"transaction created successfully",
		slog.String("id", newTxn.ID.String()),
	)

	return newTxn, nil
}

// Update modifies a transaction and recalculates allocations if needed.
// Also corrects account balances for any changes to amount or account assignments.
func (s *TransactionService) Update(
	ctx context.Context,
	id uuid.UUID,
	name string,
	category transaction.TransactionCategory,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	jarID *uuid.UUID,
	amount int64,
	occurredAt time.Time,
	isMasterIncome bool,
) (*transaction.Transaction, error) {
	s.logger.DebugContext(
		ctx,
		"updating transaction",
		slog.String("id", id.String()),
	)

	var txn *transaction.Transaction
	var oldAmount int64
	var oldIsMasterIncome bool
	var oldType transaction.TransactionType
	var oldFromAccountID *uuid.UUID
	var oldToAccountID *uuid.UUID
	var oldJarID *uuid.UUID
	var err error

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		txn, err = s.transactionRepo.FindByID(txCtx, id)
		if err != nil {
			s.logger.WarnContext(
				txCtx,
				"transaction not found for update",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		// Preserve old values for balance correction
		oldType = txn.Type
		oldAmount = txn.Amount
		oldIsMasterIncome = txn.IsMasterIncome
		oldFromAccountID = txn.FromAccountID
		oldToAccountID = txn.ToAccountID
		oldJarID = txn.JarID

		if err := txn.Update(
			name,
			category,
			fromAccountID,
			toAccountID,
			jarID,
			amount,
			occurredAt,
			isMasterIncome,
		); err != nil {
			s.logger.WarnContext(
				txCtx,
				"transaction validation failed",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		// Reverse old balance effects
		if err := s.reverseAccountBalances(txCtx, oldType, oldFromAccountID, oldToAccountID, oldAmount); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to reverse old account balances",
				slog.String("error", err.Error()),
			)
			return err
		}

		// Apply new balance effects
		if err := s.applyAccountBalances(txCtx, txn.Type, txn.FromAccountID, txn.ToAccountID, txn.Amount); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to apply new account balances",
				slog.String("error", err.Error()),
			)
			return err
		}

		if txn.HasAllocationChanged(oldAmount, oldIsMasterIncome, oldJarID) {
			if err := s.allocationRepo.DeleteByTransactionID(txCtx, id); err != nil {
				s.logger.ErrorContext(
					txCtx,
					"failed to delete old allocations",
					slog.String("error", err.Error()),
				)
				return err
			}

			if txn.IsMasterIncome {
				if err := s.createAllocations(txCtx, txn); err != nil {
					s.logger.ErrorContext(
						txCtx,
						"failed to create new allocations",
						slog.String("error", err.Error()),
					)
					return err
				}
			} else if txn.Type == transaction.TransactionTypeIncome && txn.JarID != nil {
				alloc := &transaction.Allocation{
					ID:            uuid.New(),
					TransactionID: id,
					JarID:         *txn.JarID,
					Amount:        txn.Amount,
				}

				if err := s.allocationRepo.Create(txCtx, alloc); err != nil {
					s.logger.ErrorContext(
						txCtx,
						"failed to create allocation",
						slog.String("error", err.Error()),
					)
					return err
				}
			}
		}

		if err := s.transactionRepo.Save(txCtx, txn); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to save transaction",
				slog.String("error", err.Error()),
			)
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.logger.InfoContext(
		ctx,
		"transaction updated successfully",
		slog.String("id", id.String()),
	)

	return txn, nil
}

// Delete removes a transaction and its allocations atomically.
// Also reverses the transaction's account balance effects.
func (s *TransactionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.DebugContext(
		ctx,
		"deleting transaction",
		slog.String("id", id.String()),
	)

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Fetch the transaction to get its balance-affecting fields
		txn, err := s.transactionRepo.FindByID(txCtx, id)
		if err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to find transaction for deletion",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		// Reverse account balance effects before deletion
		if err := s.reverseAccountBalances(
			txCtx,
			txn.Type,
			txn.FromAccountID,
			txn.ToAccountID,
			txn.Amount,
		); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to reverse account balances during deletion",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		if err := s.allocationRepo.DeleteByTransactionID(txCtx, id); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to delete allocations",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		if err := s.transactionRepo.Delete(txCtx, id); err != nil {
			s.logger.ErrorContext(
				txCtx,
				"failed to delete transaction",
				slog.String("id", id.String()),
				slog.String("error", err.Error()),
			)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.logger.InfoContext(
		ctx,
		"transaction deleted successfully",
		slog.String("id", id.String()),
	)

	return nil
}

// List retrieves transactions with optional filtering and pagination.
func (s *TransactionService) List(
	ctx context.Context,
	params ports.ListParams,
) ([]*ports.TransactionListItem, error) {
	s.logger.DebugContext(
		ctx,
		"listing transactions",
		slog.Int("limit", params.Limit),
		slog.Int("offset", params.Offset),
	)

	txns, err := s.transactionRepo.List(ctx, params)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to list transactions",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return txns, nil
}

// ListByAccount retrieves transactions for a specific account with running balance.
func (s *TransactionService) ListByAccount(
	ctx context.Context,
	accountID uuid.UUID,
	params ports.ListParams,
) ([]*ports.TransactionListItem, error) {
	s.logger.DebugContext(
		ctx,
		"listing transactions by account",
		slog.String("account_id", accountID.String()),
		slog.Int("limit", params.Limit),
		slog.Int("offset", params.Offset),
	)

	txns, err := s.transactionRepo.ListByAccount(ctx, accountID, params)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to list transactions by account",
			slog.String("account_id", accountID.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return txns, nil
}

// GetByID retrieves a single transaction.
func (s *TransactionService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*transaction.Transaction, error) {
	txn, err := s.transactionRepo.FindByID(ctx, id)
	if err != nil {
		s.logger.WarnContext(
			ctx,
			"transaction not found",
			slog.String("id", id.String()),
		)
		return nil, err
	}

	return txn, nil
}

// createAllocations distributes income to jars using jar.AllocateIncome.
func (s *TransactionService) createAllocations(
	ctx context.Context,
	txn *transaction.Transaction,
) error {
	allJars, err := s.jarRepo.List(ctx)
	if err != nil {
		return err
	}

	activeJars := make([]*jar.Jar, 0, len(allJars))

	for _, j := range allJars {
		if !j.IsArchived {
			activeJars = append(activeJars, j)
		}
	}

	allocations, err := jar.AllocateIncome(activeJars, txn.Amount)
	if err != nil {
		return err
	}

	for jarID, amount := range allocations {
		alloc := &transaction.Allocation{
			ID:            uuid.New(),
			TransactionID: txn.ID,
			JarID:         jarID,
			Amount:        amount,
		}

		if err := s.allocationRepo.Create(ctx, alloc); err != nil {
			return err
		}
	}

	return nil
}

func (s *TransactionService) updateAccountBalances(
	ctx context.Context,
	txn *transaction.Transaction,
) error {
	return s.applyAccountBalances(ctx, txn.Type, txn.FromAccountID, txn.ToAccountID, txn.Amount)
}

// applyAccountBalances applies the balance effects of a transaction.
// This is used during Create and as part of Update (after reversal).
func (s *TransactionService) applyAccountBalances(
	ctx context.Context,
	txnType transaction.TransactionType,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	amount int64,
) error {
	switch txnType {
	case transaction.TransactionTypeIncome:
		if toAccountID == nil {
			return transaction.ErrTargetAccountRequired
		}

		return s.accountRepo.AdjustBalance(ctx, *toAccountID, amount)

	case transaction.TransactionTypeExpense:
		if fromAccountID == nil {
			return transaction.ErrSourceAccountRequired
		}

		return s.accountRepo.AdjustBalance(ctx, *fromAccountID, -amount)

	case transaction.TransactionTypeTransfer:
		if fromAccountID == nil || toAccountID == nil {
			return transaction.ErrInvalidAccount
		}

		if err := s.accountRepo.AdjustBalance(ctx, *fromAccountID, -amount); err != nil {
			return err
		}

		return s.accountRepo.AdjustBalance(ctx, *toAccountID, amount)
	}

	return nil
}

// reverseAccountBalances reverses the balance effects of a transaction.
// This is the inverse of applyAccountBalances.
// Used during Update (to undo old effects) and Delete (to undo all effects).
func (s *TransactionService) reverseAccountBalances(
	ctx context.Context,
	txnType transaction.TransactionType,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	amount int64,
) error {
	switch txnType {
	case transaction.TransactionTypeIncome:
		if toAccountID == nil {
			return transaction.ErrTargetAccountRequired
		}

		return s.accountRepo.AdjustBalance(ctx, *toAccountID, -amount)

	case transaction.TransactionTypeExpense:
		if fromAccountID == nil {
			return transaction.ErrSourceAccountRequired
		}

		return s.accountRepo.AdjustBalance(ctx, *fromAccountID, amount)

	case transaction.TransactionTypeTransfer:
		if fromAccountID == nil || toAccountID == nil {
			return transaction.ErrInvalidAccount
		}

		if err := s.accountRepo.AdjustBalance(ctx, *fromAccountID, amount); err != nil {
			return err
		}

		return s.accountRepo.AdjustBalance(ctx, *toAccountID, -amount)
	}

	return nil
}
