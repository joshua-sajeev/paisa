package postgres_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

func setupBenchmark(b *testing.B, count int) (uuid.UUID, uuid.UUID, []uuid.UUID) {
	b.Helper()

	// Truncate tables first
	_, err := db.Exec(ctx, "TRUNCATE TABLE accounts, jars CASCADE")
	if err != nil {
		b.Fatalf("failed to truncate: %v", err)
	}

	// Create accounts
	acc1 := &account.Account{
		ID:         uuid.New(),
		Name:       "Checking",
		IsArchived: false,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	acc2 := &account.Account{
		ID:         uuid.New(),
		Name:       "Savings",
		IsArchived: false,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := accountRepo.Create(ctx, acc1); err != nil {
		b.Fatalf("failed to create acc1: %v", err)
	}
	if err := accountRepo.Create(ctx, acc2); err != nil {
		b.Fatalf("failed to create acc2: %v", err)
	}

	// Create a Jar
	j := &jar.Jar{
		ID:              uuid.New(),
		Name:            "Emergency Fund",
		AllocationType:  jar.AllocationTypeFixed,
		AllocationValue: 10000,
		IsArchived:      false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := jarRepo.Create(ctx, j); err != nil {
		b.Fatalf("failed to create jar: %v", err)
	}

	txIDs := make([]uuid.UUID, 0, count)

	// Create transactions
	for i := 0; i < count; i++ {
		var tx *transaction.Transaction
		occurred := time.Now().UTC().Add(-time.Duration(i) * time.Hour)

		switch i % 5 {
		case 0:
			// Income to acc1
			tx = &transaction.Transaction{
				ID:             uuid.New(),
				Name:           fmt.Sprintf("Salary %d", i),
				Type:           transaction.TransactionTypeIncome,
				Category:       transaction.TransactionCategoryOther,
				FromAccountID:  nil,
				ToAccountID:    &acc1.ID,
				JarID:          nil,
				Amount:         500000,
				OccurredAt:     occurred,
				IsMasterIncome: true,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}
		case 1:
			// Income to acc2
			tx = &transaction.Transaction{
				ID:             uuid.New(),
				Name:           fmt.Sprintf("Interest %d", i),
				Type:           transaction.TransactionTypeIncome,
				Category:       transaction.TransactionCategoryOther,
				FromAccountID:  nil,
				ToAccountID:    &acc2.ID,
				JarID:          nil,
				Amount:         10000,
				OccurredAt:     occurred,
				IsMasterIncome: false,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}
		case 2:
			// Expense from acc1
			tx = &transaction.Transaction{
				ID:             uuid.New(),
				Name:           fmt.Sprintf("Groceries %d", i),
				Type:           transaction.TransactionTypeExpense,
				Category:       transaction.TransactionCategoryGroceries,
				FromAccountID:  &acc1.ID,
				ToAccountID:    nil,
				JarID:          nil,
				Amount:         15000,
				OccurredAt:     occurred,
				IsMasterIncome: false,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}
		case 3:
			// Expense from acc2, with jar
			tx = &transaction.Transaction{
				ID:             uuid.New(),
				Name:           fmt.Sprintf("Health %d", i),
				Type:           transaction.TransactionTypeExpense,
				Category:       transaction.TransactionCategoryHealth,
				FromAccountID:  &acc2.ID,
				ToAccountID:    nil,
				JarID:          &j.ID,
				Amount:         20000,
				OccurredAt:     occurred,
				IsMasterIncome: false,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}
		case 4:
			// Transfer acc1 -> acc2
			tx = &transaction.Transaction{
				ID:             uuid.New(),
				Name:           fmt.Sprintf("Transfer %d", i),
				Type:           transaction.TransactionTypeTransfer,
				Category:       transaction.TransactionCategoryTransfer,
				FromAccountID:  &acc1.ID,
				ToAccountID:    &acc2.ID,
				JarID:          nil,
				Amount:         50000,
				OccurredAt:     occurred,
				IsMasterIncome: false,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			}
		}

		if err := transactionRepo.Create(ctx, tx); err != nil {
			b.Fatalf("failed to create transaction %d: %v", i, err)
		}
		txIDs = append(txIDs, tx.ID)
	}

	return acc1.ID, acc2.ID, txIDs
}

func BenchmarkTransactionRepository_List(b *testing.B) {
	_, _, _ = setupBenchmark(b, 500)

	b.Run("Limit=10,NoFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:  10,
				Offset: 0,
			}
			res, err := transactionRepo.List(ctx, params)
			if err != nil {
				b.Fatalf("List error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})

	b.Run("Limit=100,NoFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:  100,
				Offset: 0,
			}
			res, err := transactionRepo.List(ctx, params)
			if err != nil {
				b.Fatalf("List error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})

	b.Run("Limit=10,CategoryFilter", func(b *testing.B) {
		b.ResetTimer()
		cat := transaction.TransactionCategoryGroceries
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:    10,
				Offset:   0,
				Category: &cat,
			}
			res, err := transactionRepo.List(ctx, params)
			if err != nil {
				b.Fatalf("List error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})

	b.Run("Limit=10,SearchFilter", func(b *testing.B) {
		b.ResetTimer()
		search := "Groceries"
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:  10,
				Offset: 0,
				Search: &search,
			}
			res, err := transactionRepo.List(ctx, params)
			if err != nil {
				b.Fatalf("List error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})
}

func BenchmarkTransactionRepository_ListByAccount(b *testing.B) {
	acc1ID, _, _ := setupBenchmark(b, 500)

	b.Run("Limit=10,NoFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:  10,
				Offset: 0,
			}
			res, err := transactionRepo.ListByAccount(ctx, acc1ID, params)
			if err != nil {
				b.Fatalf("ListByAccount error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})

	b.Run("Limit=100,NoFilter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:  100,
				Offset: 0,
			}
			res, err := transactionRepo.ListByAccount(ctx, acc1ID, params)
			if err != nil {
				b.Fatalf("ListByAccount error: %v", err)
			}
			if len(res) == 0 {
				b.Fatal("expected transactions, got 0")
			}
		}
	})
}

func BenchmarkTransactionRepository_FindByID(b *testing.B) {
	_, _, txIDs := setupBenchmark(b, 100)
	if len(txIDs) == 0 {
		b.Fatal("no transactions created")
	}
	targetID := txIDs[0]
	nonExistID := uuid.New()

	b.Run("Existing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			res, err := transactionRepo.FindByID(ctx, targetID)
			if err != nil {
				b.Fatalf("FindByID error: %v", err)
			}
			if res == nil {
				b.Fatal("expected transaction, got nil")
			}
		}
	})

	b.Run("NonExisting", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := transactionRepo.FindByID(ctx, nonExistID)
			if err == nil {
				b.Fatal("expected error, got nil")
			}
		}
	})
}
