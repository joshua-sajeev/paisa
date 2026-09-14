package postgres_test

import (
	"testing"
	"time"

	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/seed"
)

func TestDashboardRecentTransactionsProjectsDisplayFields(t *testing.T) {
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

	got, err := dashboardRepo.GetRecentTransactions(ctx, 10)
	if err != nil {
		t.Fatalf("GetRecentTransactions() error = %v", err)
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

func BenchmarkDashboardRepository(b *testing.B) {
	if err := seed.Run(ctx, db); err != nil {
		b.Fatalf("failed to seed benchmark data: %v", err)
	}

	b.Run("GetTotalBalance", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := dashboardRepo.GetTotalBalance(ctx)
			if err != nil {
				b.Fatalf("GetTotalBalance error: %v", err)
			}
		}
	})

	b.Run("GetMonthlySummary", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err := dashboardRepo.GetMonthlySummary(ctx)
			if err != nil {
				b.Fatalf("GetMonthlySummary error: %v", err)
			}
			if res == nil {
				b.Fatal("expected summary, got nil")
			}
		}
	})

	b.Run("GetAccountBalances", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err := dashboardRepo.GetAccountBalances(ctx)
			if err != nil {
				b.Fatalf("GetAccountBalances error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected accounts, got 0")
			}
		}
	})

	b.Run("GetJarSummaries", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err := dashboardRepo.GetJarSummaries(ctx)
			if err != nil {
				b.Fatalf("GetJarSummaries error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected jars, got 0")
			}
		}
	})

	b.Run("GetGoalSummaries", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err := dashboardRepo.GetGoalSummaries(ctx)
			if err != nil {
				b.Fatalf("GetGoalSummaries error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected goals, got 0")
			}
		}
	})

	b.Run("GetRecentTransactions", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			res, err := dashboardRepo.GetRecentTransactions(ctx, 10)
			if err != nil {
				b.Fatalf("GetRecentTransactions error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})
}
