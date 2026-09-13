package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/seed"
)

func setupBenchmark(b *testing.B, count int) (uuid.UUID, uuid.UUID, []uuid.UUID) {
	b.Helper()

	if err := seed.RunWithCount(ctx, db, count); err != nil {
		b.Fatalf("failed to seed: %v", err)
	}

	// Fetch accounts from DB to return active IDs
	var accIDs []uuid.UUID
	rows, err := db.Query(ctx, "SELECT id FROM accounts ORDER BY name")
	if err != nil {
		b.Fatalf("failed to query accounts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			b.Fatalf("failed to scan account id: %v", err)
		}
		accIDs = append(accIDs, id)
	}
	if len(accIDs) < 2 {
		b.Fatalf("expected at least 2 accounts, got %d", len(accIDs))
	}

	// Fetch transactions from DB to return active IDs
	var txIDs []uuid.UUID
	txRows, err := db.Query(ctx, "SELECT id FROM transactions LIMIT $1", count)
	if err != nil {
		b.Fatalf("failed to query transactions: %v", err)
	}
	defer txRows.Close()

	for txRows.Next() {
		var id uuid.UUID
		if err := txRows.Scan(&id); err != nil {
			b.Fatalf("failed to scan transaction id: %v", err)
		}
		txIDs = append(txIDs, id)
	}

	return accIDs[0], accIDs[1], txIDs
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
