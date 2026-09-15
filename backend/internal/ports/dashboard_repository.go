package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DashboardSummary contains monthly financial aggregates
type DashboardSummary struct {
	TotalBalance         int64
	MonthlyIncome        int64
	MonthlyExpense       int64
	MonthlySavings       int64
	MonthlySavingsChange float64
}

// AccountBalance represents an account's balance
type AccountBalance struct {
	ID      uuid.UUID
	Name    string
	Balance int64
}

// JarSummary represents jar allocation and spending
type JarSummary struct {
	ID                  uuid.UUID
	Name                string
	AllocationType      string
	AllocationValue     int64
	Allocated           int64
	Used                int64
	Available           int64
	UsedPercentage      float64
	AvailablePercentage float64
}

// GoalSummary represents goal progress
type GoalSummary struct {
	ID          uuid.UUID
	Name        string
	Target      int64
	Contributed int64
	Remaining   int64
	Progress    float64
	Deadline    time.Time
}

// DashboardRepository defines dashboard data access operations
type DashboardRepository interface {
	GetTotalBalance(ctx context.Context) (int64, error)
	GetMonthlySummary(ctx context.Context) (*DashboardSummary, error)
	GetAccountBalances(ctx context.Context) ([]*AccountBalance, error)
	GetJarSummaries(ctx context.Context) ([]*JarSummary, error)
	GetGoalSummaries(ctx context.Context) ([]*GoalSummary, error)
	GetRecentTransactions(ctx context.Context, limit int) ([]*TransactionListItem, error)
}
