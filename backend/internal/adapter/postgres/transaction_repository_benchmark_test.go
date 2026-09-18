package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

func BenchmarkTransactionRepository_List(b *testing.B) {
	_, _, _ = setupBenchmark(b, 10000)

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
	acc1ID, _, _ := setupBenchmark(b, 10000)

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

	b.Run("Limit=10,ComplexFilters", func(b *testing.B) {
		b.ResetTimer()
		cat := transaction.TransactionCategoryGroceries
		search := "Groceries"
		for i := 0; i < b.N; i++ {
			params := ports.ListParams{
				Limit:    10,
				Offset:   0,
				Category: &cat,
				Search:   &search,
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
	_, _, txIDs := setupBenchmark(b, 10000)
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
