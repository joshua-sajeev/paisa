package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
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
	return mustCreateTransactionWithJar(
		t,
		name,
		transactionType,
		category,
		fromAccountID,
		toAccountID,
		nil,
		amount,
		occurredAt,
	)
}

func mustCreateTransactionWithJar(
	t *testing.T,
	name string,
	transactionType transaction.TransactionType,
	category transaction.TransactionCategory,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	jarID *uuid.UUID,
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
		jarID,
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

func findTransactionListItem(
	t *testing.T,
	items []*ports.TransactionListItem,
	id uuid.UUID,
) *ports.TransactionListItem {
	t.Helper()

	for _, item := range items {
		if item.ID == id {
			return item
		}
	}

	t.Fatalf("transaction list item %v not found", id)
	return nil
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

func TestTransactionListProjectsGlobalDisplayFields(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	checking := newTestAccount("Checking")
	checking.Balance = 750
	savings := newTestAccount("Savings")
	savings.Balance = 1200

	for _, acc := range []*account.Account{checking, savings} {
		if err := accountRepo.Create(ctx, acc); err != nil {
			t.Fatalf("Create() account error = %v", err)
		}
	}

	needs := newTestJar(
		"Needs",
		jar.AllocationTypePercentage,
		50,
	)
	if err := jarRepo.Create(ctx, needs); err != nil {
		t.Fatalf("Create() jar error = %v", err)
	}

	jan1 := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	jan2 := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)
	jan3 := time.Date(2026, 1, 3, 9, 0, 0, 0, time.UTC)

	income := mustCreateTransaction(
		t,
		"Salary",
		transaction.TransactionTypeIncome,
		transaction.TransactionCategoryOther,
		nil,
		&checking.ID,
		1000,
		jan1,
	)
	expense := mustCreateTransactionWithJar(
		t,
		"Groceries",
		transaction.TransactionTypeExpense,
		transaction.TransactionCategoryGroceries,
		&checking.ID,
		nil,
		&needs.ID,
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

	got, err := transactionRepo.List(
		ctx,
		ports.ListParams{Limit: 10, Offset: 0},
	)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	incomeItem := findTransactionListItem(t, got, income.ID)
	if incomeItem.Account != checking.Name {
		t.Errorf("income Account = %q, want %q", incomeItem.Account, checking.Name)
	}
	if incomeItem.AccountBalance == nil || *incomeItem.AccountBalance != checking.Balance {
		t.Errorf("income AccountBalance = %v, want %d", incomeItem.AccountBalance, checking.Balance)
	}

	expenseItem := findTransactionListItem(t, got, expense.ID)
	if expenseItem.Account != checking.Name {
		t.Errorf("expense Account = %q, want %q", expenseItem.Account, checking.Name)
	}
	if expenseItem.AccountBalance == nil || *expenseItem.AccountBalance != checking.Balance {
		t.Errorf("expense AccountBalance = %v, want %d", expenseItem.AccountBalance, checking.Balance)
	}
	if expenseItem.JarName == nil || *expenseItem.JarName != needs.Name {
		t.Errorf("expense JarName = %v, want %q", expenseItem.JarName, needs.Name)
	}

	transferItem := findTransactionListItem(t, got, transfer.ID)
	wantTransferAccount := checking.Name + " -> " + savings.Name
	if transferItem.Account != wantTransferAccount {
		t.Errorf("transfer Account = %q, want %q", transferItem.Account, wantTransferAccount)
	}
	if transferItem.AccountBalance != nil {
		t.Errorf("transfer AccountBalance = %v, want nil", transferItem.AccountBalance)
	}
	if transferItem.Amount != transfer.Amount {
		t.Errorf("transfer Amount = %d, want %d", transferItem.Amount, transfer.Amount)
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

	if got[0].ID != income.ID {
		t.Errorf("first transaction ID = %v, want %v", got[0].ID, income.ID)
	}

	if got[0].AccountBalance == nil || *got[0].AccountBalance != 550 {
		t.Errorf("first AccountBalance = %v, want %d", got[0].AccountBalance, 550)
	}

	if got[1].ID != transfer.ID {
		t.Errorf("second transaction ID = %v, want %v", got[1].ID, transfer.ID)
	}

	if got[1].AccountBalance == nil || *got[1].AccountBalance != 500 {
		t.Errorf("second AccountBalance = %v, want %d", got[1].AccountBalance, 500)
	}

	if got[1].Amount != -300 {
		t.Errorf("transfer-out Amount = %d, want %d", got[1].Amount, -300)
	}

	if got[1].Account != savings.Name {
		t.Errorf("transfer-out Account = %q, want %q", got[1].Account, savings.Name)
	}

	gotSavings, err := transactionRepo.ListByAccount(
		ctx,
		savings.ID,
		ports.ListParams{Limit: 10, Offset: 0},
	)
	if err != nil {
		t.Fatalf("ListByAccount() savings error = %v", err)
	}

	if len(gotSavings) != 1 {
		t.Fatalf("ListByAccount() savings returned %d transactions, want 1", len(gotSavings))
	}

	if gotSavings[0].ID != transfer.ID {
		t.Errorf("savings transaction ID = %v, want %v", gotSavings[0].ID, transfer.ID)
	}

	if gotSavings[0].Amount != 300 {
		t.Errorf("transfer-in Amount = %d, want %d", gotSavings[0].Amount, 300)
	}

	if gotSavings[0].Account != checking.Name {
		t.Errorf("transfer-in Account = %q, want %q", gotSavings[0].Account, checking.Name)
	}
}

func TestTransactionListByAccountRunningBalanceSurvivesPaginationAndDateFilters(t *testing.T) {
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
	expense := mustCreateTransaction(
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
	mustCreateTransaction(
		t,
		"Refund",
		transaction.TransactionTypeIncome,
		transaction.TransactionCategoryOther,
		nil,
		&checking.ID,
		50,
		jan4,
	)

	paged, err := transactionRepo.ListByAccount(
		ctx,
		checking.ID,
		ports.ListParams{Limit: 1, Offset: 1},
	)
	if err != nil {
		t.Fatalf("ListByAccount() paged error = %v", err)
	}

	if len(paged) != 1 {
		t.Fatalf("paged ListByAccount() returned %d transactions, want 1", len(paged))
	}

	if paged[0].ID != transfer.ID {
		t.Errorf("paged transaction ID = %v, want %v", paged[0].ID, transfer.ID)
	}

	if paged[0].AccountBalance == nil || *paged[0].AccountBalance != 500 {
		t.Errorf("paged AccountBalance = %v, want %d", paged[0].AccountBalance, 500)
	}

	fromDate := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	filtered, err := transactionRepo.ListByAccount(
		ctx,
		checking.ID,
		ports.ListParams{
			Limit:    10,
			Offset:   0,
			FromDate: &fromDate,
			ToDate:   &toDate,
		},
	)
	if err != nil {
		t.Fatalf("ListByAccount() filtered error = %v", err)
	}

	if len(filtered) != 2 {
		t.Fatalf("filtered ListByAccount() returned %d transactions, want 2", len(filtered))
	}

	if filtered[0].ID != transfer.ID {
		t.Errorf("first filtered transaction ID = %v, want %v", filtered[0].ID, transfer.ID)
	}
	if filtered[0].AccountBalance == nil || *filtered[0].AccountBalance != 500 {
		t.Errorf("first filtered AccountBalance = %v, want %d", filtered[0].AccountBalance, 500)
	}

	if filtered[1].ID != expense.ID {
		t.Errorf("second filtered transaction ID = %v, want %v", filtered[1].ID, expense.ID)
	}
	if filtered[1].AccountBalance == nil || *filtered[1].AccountBalance != 800 {
		t.Errorf("second filtered AccountBalance = %v, want %d", filtered[1].AccountBalance, 800)
	}
}
