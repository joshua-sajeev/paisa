package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

func mustCreateTransaction(
	t *testing.T,
	name string,
	transactionType transaction.TransactionType,
	category transaction.TransactionCategory,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	amount int64,
	occurredAt time.Time,
) *transaction.Transaction {
	t.Helper()

	txn, err := transaction.NewTransaction(
		name,
		transactionType,
		category,
		fromAccountID,
		toAccountID,
		nil,
		amount,
		occurredAt,
		false,
	)
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	txn.CreatedAt = occurredAt
	txn.UpdatedAt = occurredAt

	if err := transactionRepo.Create(ctx, txn); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	return txn
}

func TestTransactionsTableDoesNotPersistBalanceAfter(t *testing.T) {
	var exists bool

	err := db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_name = 'transactions'
			  AND column_name = 'balance_after'
		)
		`,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("query schema: %v", err)
	}

	if exists {
		t.Fatal("transactions.balance_after column exists, want computed-only balance")
	}
}

func TestTransactionListByAccountRunningBalanceIncludesHistoryBeforeFromDate(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	checking := newTestAccount("Checking")
	savings := newTestAccount("Savings")

	for _, acc := range []*account.Account{checking, savings} {
		if err := accountRepo.Create(ctx, acc); err != nil {
			t.Fatalf("Create() account error = %v", err)
		}
	}

	jan1 := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	jan2 := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	jan3 := time.Date(2026, 1, 3, 9, 0, 0, 0, time.UTC)
	jan4 := time.Date(2026, 1, 4, 9, 0, 0, 0, time.UTC)

	mustCreateTransaction(
		t,
		"Salary",
		transaction.TransactionTypeIncome,
		transaction.TransactionCategoryOther,
		nil,
		&checking.ID,
		1000,
		jan1,
	)
	mustCreateTransaction(
		t,
		"Groceries",
		transaction.TransactionTypeExpense,
		transaction.TransactionCategoryGroceries,
		&checking.ID,
		nil,
		200,
		jan2,
	)
	transfer := mustCreateTransaction(
		t,
		"Move to savings",
		transaction.TransactionTypeTransfer,
		transaction.TransactionCategoryTransfer,
		&checking.ID,
		&savings.ID,
		300,
		jan3,
	)
	income := mustCreateTransaction(
		t,
		"Refund",
		transaction.TransactionTypeIncome,
		transaction.TransactionCategoryOther,
		nil,
		&checking.ID,
		50,
		jan4,
	)

	fromDate := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	got, err := transactionRepo.ListByAccount(
		ctx,
		checking.ID,
		ports.ListParams{
			Limit:    10,
			Offset:   0,
			FromDate: &fromDate,
		},
	)
	if err != nil {
		t.Fatalf("ListByAccount() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("ListByAccount() returned %d transactions, want 2", len(got))
	}

	if got[0].Transaction.ID != income.ID {
		t.Errorf("first transaction ID = %v, want %v", got[0].Transaction.ID, income.ID)
	}

	if got[0].BalanceAfter != 550 {
		t.Errorf("first BalanceAfter = %d, want %d", got[0].BalanceAfter, 550)
	}

	if got[1].Transaction.ID != transfer.ID {
		t.Errorf("second transaction ID = %v, want %v", got[1].Transaction.ID, transfer.ID)
	}

	if got[1].BalanceAfter != 500 {
		t.Errorf("second BalanceAfter = %d, want %d", got[1].BalanceAfter, 500)
	}
}
