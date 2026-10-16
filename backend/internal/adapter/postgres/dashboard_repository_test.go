package postgres_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/seed"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
)

func mustCreateTransaction(
	t *testing.T,
	name string,
	txnType transaction.TransactionType,
	cat transaction.TransactionCategory,
	fromAcc *uuid.UUID,
	toAcc *uuid.UUID,
	amount int64,
	occurredAt time.Time,
) *transaction.Transaction {
	t.Helper()

	return mustCreateTransactionWithJar(
		t,
		name,
		txnType,
		cat,
		fromAcc,
		toAcc,
		nil,
		amount,
		occurredAt,
	)
}

func mustCreateTransactionWithJar(
	t *testing.T,
	name string,
	txnType transaction.TransactionType,
	cat transaction.TransactionCategory,
	fromAcc *uuid.UUID,
	toAcc *uuid.UUID,
	jarID *uuid.UUID,
	amount int64,
	occurredAt time.Time,
) *transaction.Transaction {
	t.Helper()
	txn, err := transaction.NewTransaction(
		name,
		txnType,
		cat,
		fromAcc,
		toAcc,
		jarID,
		amount,
		occurredAt,
		false,
	)
	if err != nil {
		t.Fatalf("NewTransaction error: %v", err)
	}
	if err := transactionRepo.Create(ctx, txn); err != nil {
		t.Fatalf("transactionRepo.Create error: %v", err)
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
	t.Fatalf("transaction item with id %v not found", id)
	return nil
}

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
		t.Errorf(
			"income Account = %q, want %q",
			incomeItem.Account,
			checking.Name,
		)
	}
	if incomeItem.AccountBalance == nil || *incomeItem.AccountBalance != checking.Balance {
		t.Errorf(
			"income AccountBalance = %v, want %d",
			incomeItem.AccountBalance,
			checking.Balance,
		)
	}

	expenseItem := findTransactionListItem(t, got, expense.ID)
	if expenseItem.Account != checking.Name {
		t.Errorf(
			"expense Account = %q, want %q",
			expenseItem.Account,
			checking.Name,
		)
	}
	if expenseItem.AccountBalance == nil || *expenseItem.AccountBalance != checking.Balance {
		t.Errorf(
			"expense AccountBalance = %v, want %d",
			expenseItem.AccountBalance,
			checking.Balance,
		)
	}
	if expenseItem.JarName == nil || *expenseItem.JarName != needs.Name {
		t.Errorf(
			"expense JarName = %v, want %q",
			expenseItem.JarName,
			needs.Name,
		)
	}

	transferItem := findTransactionListItem(t, got, transfer.ID)
	wantTransferAccount := checking.Name + " -> " + savings.Name
	if transferItem.Account != wantTransferAccount {
		t.Errorf(
			"transfer Account = %q, want %q",
			transferItem.Account,
			wantTransferAccount,
		)
	}
	if transferItem.AccountBalance != nil {
		t.Errorf(
			"transfer AccountBalance = %v, want nil",
			transferItem.AccountBalance,
		)
	}
	if transferItem.Amount != transfer.Amount {
		t.Errorf(
			"transfer Amount = %d, want %d",
			transferItem.Amount,
			transfer.Amount,
		)
	}
}

func TestJarRolloverBalancesUseISTMonthBoundaries(t *testing.T) {
	t.Cleanup(func() { truncateTables(t, ctx, db) })

	account := newTestAccount("Rollover account")
	if err := accountRepo.Create(ctx, account); err != nil {
		t.Fatalf("create account: %v", err)
	}

	rolloverJar := newTestJar("Rollover", jar.AllocationTypePercentage, 50)
	rolloverJar.Rollover = true
	leisureJar := newTestJar("Leisure", jar.AllocationTypePercentage, 50)
	leisureJar.Rollover = false
	for _, j := range []*jar.Jar{rolloverJar, leisureJar} {
		if err := jarRepo.Create(ctx, j); err != nil {
			t.Fatalf("create jar: %v", err)
		}
	}

	now := time.Now().In(timeutil.IST)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, timeutil.IST)
	monthEnd := monthStart.AddDate(0, 1, 0)
	olderMonth := monthStart.AddDate(0, -2, 0)

	addHistoricalAllocation := func(j *jar.Jar, name string, amount int64) {
		t.Helper()
		income := mustCreateTransaction(t, name, transaction.TransactionTypeIncome,
			transaction.TransactionCategoryOther, nil, &account.ID, amount, olderMonth.Add(time.Hour))
		if err := allocationRepo.Create(ctx, &transaction.Allocation{
			ID: uuid.New(), TransactionID: income.ID, JarID: j.ID, Amount: amount,
		}); err != nil {
			t.Fatalf("create allocation: %v", err)
		}
	}
	addHistoricalAllocation(rolloverJar, "Rollover allocation", 1000)
	addHistoricalAllocation(leisureJar, "Leisure allocation", 700)

	// Both jars have older activity, followed by a complete month with no activity.
	mustCreateTransactionWithJar(t, "Rollover overspend", transaction.TransactionTypeExpense,
		transaction.TransactionCategoryOther, &account.ID, nil, &rolloverJar.ID, 1200, olderMonth.Add(2*time.Hour))
	mustCreateTransactionWithJar(t, "Leisure leftover", transaction.TransactionTypeExpense,
		transaction.TransactionCategoryOther, &account.ID, nil, &leisureJar.ID, 200, olderMonth.Add(2*time.Hour))

	// An expense exactly at 00:00 IST belongs to the new month.
	mustCreateTransactionWithJar(t, "Boundary expense", transaction.TransactionTypeExpense,
		transaction.TransactionCategoryOther, &account.ID, nil, &rolloverJar.ID, 50, monthStart)

	dashboardJars, err := dashboardRepo.GetJarSummaries(ctx)
	if err != nil {
		t.Fatalf("GetJarSummaries() error = %v", err)
	}
	service := application.NewJarService(jarRepo, allocationRepo, transactionRepo, slog.Default())
	listedJars, allocated, used, carry, err := service.ListWithStats(ctx, &monthStart, &monthEnd)
	if err != nil {
		t.Fatalf("ListWithStats() error = %v", err)
	}

	byID := make(map[uuid.UUID]*ports.JarSummary, len(dashboardJars))
	for _, summary := range dashboardJars {
		byID[summary.ID] = summary
	}
	for _, j := range listedJars {
		summary := byID[j.ID]
		if summary == nil {
			t.Fatalf("dashboard summary missing jar %s", j.ID)
		}
		wantCarry, wantUsed, wantAvailable := int64(-200), int64(50), int64(-250)
		if j.ID == leisureJar.ID {
			wantCarry, wantUsed, wantAvailable = 0, 0, 0
		}
		if summary.CarryIn != wantCarry || carry[j.ID] != wantCarry {
			t.Errorf("jar %q carry-in: dashboard=%d list=%d, want %d", j.Name, summary.CarryIn, carry[j.ID], wantCarry)
		}
		if summary.Allocated != allocated[j.ID] || summary.Used != wantUsed || summary.Used != used[j.ID] || summary.Available != wantAvailable {
			t.Errorf("jar %q totals disagree or are incorrect: dashboard=(%d,%d,%d) list=(%d,%d), want available %d", j.Name, summary.Allocated, summary.Used, summary.Available, allocated[j.ID], used[j.ID], wantAvailable)
		}
	}
}

func BenchmarkDashboardRepository(b *testing.B) {
	if err := seed.RunWithCount(ctx, db, 10000); err != nil {
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

func TestDashboardGetJarSummariesIncludesIconKey(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	custom := newTestJar("Savings", jar.AllocationTypePercentage, 20)
	custom.IconKey = "piggy"

	if err := jarRepo.Create(ctx, custom); err != nil {
		t.Fatalf("Create() custom jar error = %v", err)
	}

	def := newTestJar("Needs", jar.AllocationTypeRemainder, 0)

	if err := jarRepo.Create(ctx, def); err != nil {
		t.Fatalf("Create() default jar error = %v", err)
	}

	got, err := dashboardRepo.GetJarSummaries(ctx)
	if err != nil {
		t.Fatalf("GetJarSummaries() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(got))
	}

	icons := make(map[uuid.UUID]string, len(got))
	for _, s := range got {
		icons[s.ID] = s.IconKey
	}

	if icons[custom.ID] != "piggy" {
		t.Errorf("custom IconKey = %q, want %q", icons[custom.ID], "piggy")
	}

	if icons[def.ID] != jar.DefaultIconKey {
		t.Errorf("default IconKey = %q, want %q", icons[def.ID], jar.DefaultIconKey)
	}
}
