package postgres_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

func setupTestTransactionService(t *testing.T) *application.TransactionService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return application.NewTransactionService(
		transactionRepo,
		allocationRepo,
		jarRepo,
		accountRepo,
		txManager,
		logger,
	)
}

type dbAllocation struct {
	JarID  uuid.UUID
	Amount int64
}

func getDBAllocations(t *testing.T, ctx context.Context, transactionID uuid.UUID) []dbAllocation {
	t.Helper()

	rows, err := db.Query(
		ctx,
		"SELECT jar_id, amount FROM jar_allocations WHERE transaction_id = $1 ORDER BY jar_id",
		transactionID,
	)
	if err != nil {
		t.Fatalf("failed to query jar_allocations: %v", err)
	}
	defer rows.Close()

	var allocs []dbAllocation
	for rows.Next() {
		var a dbAllocation
		if err := rows.Scan(&a.JarID, &a.Amount); err != nil {
			t.Fatalf("failed to scan jar_allocations: %v", err)
		}
		allocs = append(allocs, a)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("jar_allocations rows error: %v", err)
	}

	return allocs
}

func getDBTransactionJarID(t *testing.T, ctx context.Context, transactionID uuid.UUID) *uuid.UUID {
	t.Helper()

	var jarID *uuid.UUID
	err := db.QueryRow(
		ctx,
		"SELECT jar_id FROM transactions WHERE id = $1",
		transactionID,
	).Scan(&jarID)
	if err != nil {
		t.Fatalf("failed to query transactions.jar_id: %v", err)
	}

	return jarID
}

func findJarSummary(t *testing.T, summaries []*ports.JarSummary, jarID uuid.UUID) *ports.JarSummary {
	t.Helper()
	for _, s := range summaries {
		if s.ID == jarID {
			return s
		}
	}
	t.Fatalf("jar summary for jar ID %v not found", jarID)
	return nil
}

func TestJarExpenseAllocationAndReassignment(t *testing.T) {
	truncateTables(t, ctx, db)

	svc := setupTestTransactionService(t)

	// Create test account
	acc, err := account.NewAccount("Main Account")
	if err != nil {
		t.Fatalf("NewAccount error: %v", err)
	}
	if err := accountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("AccountRepo.Create error: %v", err)
	}

	// Create test jars: Jar A (Fixed 85000), Jar Remainder (Remainder)
	jarA := newTestJar("Necessities", jar.AllocationTypeFixed, 85000)
	if err := jarRepo.Create(ctx, jarA); err != nil {
		t.Fatalf("JarRepo.Create Jar A error: %v", err)
	}

	jarRemainder := newTestJar("Savings", jar.AllocationTypeRemainder, 0)
	if err := jarRepo.Create(ctx, jarRemainder); err != nil {
		t.Fatalf("JarRepo.Create Jar Remainder error: %v", err)
	}

	jarB := newTestJar("Leisure", jar.AllocationTypeFixed, 10000)
	if err := jarRepo.Create(ctx, jarB); err != nil {
		t.Fatalf("JarRepo.Create Jar B error: %v", err)
	}

	occurredAt := time.Now().UTC()

	// 1. Master income allocation
	t.Run("master income allocation creates jar_allocations", func(t *testing.T) {
		masterTx, err := svc.Create(
			ctx,
			"Monthly Salary",
			transaction.TransactionTypeIncome,
			transaction.TransactionCategoryOther,
			nil,
			&acc.ID,
			nil,
			100000,
			occurredAt,
			true,
		)
		if err != nil {
			t.Fatalf("Create master income error: %v", err)
		}

		allocs := getDBAllocations(t, ctx, masterTx.ID)
		if len(allocs) < 2 {
			t.Fatalf("master income expected at least 2 jar_allocations, got %d", len(allocs))
		}

		summaries, err := dashboardRepo.GetJarSummaries(ctx)
		if err != nil {
			t.Fatalf("GetJarSummaries error: %v", err)
		}

		sumA := findJarSummary(t, summaries, jarA.ID)
		if sumA.Allocated != 85000 {
			t.Errorf("Jar A Allocated = %d, want 85000", sumA.Allocated)
		}
		if sumA.Used != 0 {
			t.Errorf("Jar A Used = %d, want 0", sumA.Used)
		}
		if sumA.Available != 85000 {
			t.Errorf("Jar A Available = %d, want 85000", sumA.Available)
		}
	})

	// 2. Normal expense with jar does NOT create a jar_allocations row, but increases used
	var expenseTx *transaction.Transaction
	t.Run("normal expense with jar", func(t *testing.T) {
		var err error
		expenseTx, err = svc.Create(
			ctx,
			"Groceries",
			transaction.TransactionTypeExpense,
			transaction.TransactionCategoryGroceries,
			&acc.ID,
			nil,
			&jarA.ID,
			10000,
			occurredAt,
			false,
		)
		if err != nil {
			t.Fatalf("Create expense error: %v", err)
		}

		// Verify 0 rows in jar_allocations for normal expense
		allocs := getDBAllocations(t, ctx, expenseTx.ID)
		if len(allocs) != 0 {
			t.Fatalf("normal expense must NOT create jar_allocations, got %d rows", len(allocs))
		}

		summaries, err := dashboardRepo.GetJarSummaries(ctx)
		if err != nil {
			t.Fatalf("GetJarSummaries error: %v", err)
		}

		sumA := findJarSummary(t, summaries, jarA.ID)
		if sumA.Allocated != 85000 {
			t.Errorf("Jar A Allocated = %d, want 85000 (unchanged)", sumA.Allocated)
		}
		if sumA.Used != 10000 {
			t.Errorf("Jar A Used = %d, want 10000", sumA.Used)
		}
		if sumA.Available != 75000 {
			t.Errorf("Jar A Available = %d, want 75000 (85000 - 10000)", sumA.Available)
		}
	})

	// 3. Normal expense without jar
	t.Run("normal expense without jar", func(t *testing.T) {
		noJarTx, err := svc.Create(
			ctx,
			"Cash Outlay",
			transaction.TransactionTypeExpense,
			transaction.TransactionCategoryOther,
			&acc.ID,
			nil,
			nil,
			5000,
			occurredAt,
			false,
		)
		if err != nil {
			t.Fatalf("Create expense without jar error: %v", err)
		}

		allocs := getDBAllocations(t, ctx, noJarTx.ID)
		if len(allocs) != 0 {
			t.Fatalf("expense without jar must NOT create jar_allocations, got %d rows", len(allocs))
		}
	})

	// 4. Changing expense amount (10000 -> 25000)
	t.Run("changing expense amount", func(t *testing.T) {
		_, err := svc.Update(
			ctx,
			expenseTx.ID,
			"Groceries Updated",
			transaction.TransactionCategoryGroceries,
			&acc.ID,
			nil,
			&jarA.ID,
			25000,
			occurredAt,
			false,
		)
		if err != nil {
			t.Fatalf("Update expense amount error: %v", err)
		}

		allocs := getDBAllocations(t, ctx, expenseTx.ID)
		if len(allocs) != 0 {
			t.Fatalf("updated expense must NOT create jar_allocations, got %d rows", len(allocs))
		}

		summaries, err := dashboardRepo.GetJarSummaries(ctx)
		if err != nil {
			t.Fatalf("GetJarSummaries error: %v", err)
		}

		sumA := findJarSummary(t, summaries, jarA.ID)
		if sumA.Allocated != 85000 {
			t.Errorf("Jar A Allocated = %d, want 85000", sumA.Allocated)
		}
		if sumA.Used != 25000 {
			t.Errorf("Jar A Used = %d, want 25000", sumA.Used)
		}
		if sumA.Available != 60000 {
			t.Errorf("Jar A Available = %d, want 60000", sumA.Available)
		}
	})

	// 5. Changing expense jar A -> jar B
	t.Run("changing expense jar A to jar B", func(t *testing.T) {
		_, err := svc.Update(
			ctx,
			expenseTx.ID,
			"Groceries Moved to Jar B",
			transaction.TransactionCategoryGroceries,
			&acc.ID,
			nil,
			&jarB.ID,
			25000,
			occurredAt,
			false,
		)
		if err != nil {
			t.Fatalf("Update expense jar error: %v", err)
		}

		allocs := getDBAllocations(t, ctx, expenseTx.ID)
		if len(allocs) != 0 {
			t.Fatalf("reassigned expense must NOT create jar_allocations, got %d rows", len(allocs))
		}

		summaries, err := dashboardRepo.GetJarSummaries(ctx)
		if err != nil {
			t.Fatalf("GetJarSummaries error: %v", err)
		}

		sumA := findJarSummary(t, summaries, jarA.ID)
		if sumA.Used != 0 {
			t.Errorf("Jar A Used = %d, want 0 (moved to Jar B)", sumA.Used)
		}
		if sumA.Available != 85000 {
			t.Errorf("Jar A Available = %d, want 85000", sumA.Available)
		}

		sumB := findJarSummary(t, summaries, jarB.ID)
		if sumB.Used != 25000 {
			t.Errorf("Jar B Used = %d, want 25000", sumB.Used)
		}
	})

	// 6. Expense deletion
	t.Run("expense deletion", func(t *testing.T) {
		if err := svc.Delete(ctx, expenseTx.ID); err != nil {
			t.Fatalf("Delete expense error: %v", err)
		}

		summaries, err := dashboardRepo.GetJarSummaries(ctx)
		if err != nil {
			t.Fatalf("GetJarSummaries error: %v", err)
		}

		sumB := findJarSummary(t, summaries, jarB.ID)
		if sumB.Used != 0 {
			t.Errorf("Jar B Used = %d, want 0 after deletion", sumB.Used)
		}
	})

	// 7. Direct income to jar (e.g. bonus to Jar B) creates single jar_allocations row
	t.Run("direct income to jar and reassignment", func(t *testing.T) {
		incomeTx, err := svc.Create(
			ctx,
			"Direct Bonus to Jar A",
			transaction.TransactionTypeIncome,
			transaction.TransactionCategoryOther,
			nil,
			&acc.ID,
			&jarA.ID,
			5000,
			occurredAt,
			false,
		)
		if err != nil {
			t.Fatalf("Create direct income error: %v", err)
		}

		allocs := getDBAllocations(t, ctx, incomeTx.ID)
		if len(allocs) != 1 || allocs[0].JarID != jarA.ID || allocs[0].Amount != 5000 {
			t.Fatalf("direct income expected 1 allocation for Jar A, got %+v", allocs)
		}

		// Reassign direct income from Jar A to Jar B
		_, err = svc.Update(
			ctx,
			incomeTx.ID,
			"Direct Bonus to Jar B",
			transaction.TransactionCategoryOther,
			nil,
			&acc.ID,
			&jarB.ID,
			5000,
			occurredAt,
			false,
		)
		if err != nil {
			t.Fatalf("Update direct income jar error: %v", err)
		}

		allocs = getDBAllocations(t, ctx, incomeTx.ID)
		if len(allocs) != 1 || allocs[0].JarID != jarB.ID || allocs[0].Amount != 5000 {
			t.Fatalf("reassigned direct income expected 1 allocation for Jar B, got %+v", allocs)
		}
	})
}
