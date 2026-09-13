package postgres_test

import (
	"testing"

	"github.com/joshu-sajeev/paisa/internal/seed"
)

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
