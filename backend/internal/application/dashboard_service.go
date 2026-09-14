package application

import (
	"context"
	"log/slog"

	"github.com/joshu-sajeev/paisa/internal/ports"
)

// DashboardService handles dashboard use cases
type DashboardService struct {
	dashboardRepo ports.DashboardRepository
	logger        *slog.Logger
}

// NewDashboardService creates a new DashboardService
func NewDashboardService(
	dashboardRepo ports.DashboardRepository,
	logger *slog.Logger,
) *DashboardService {
	return &DashboardService{
		dashboardRepo: dashboardRepo,
		logger:        logger,
	}
}

// DashboardResponse is the complete dashboard response
type DashboardResponse struct {
	Summary            *ports.DashboardSummary
	TotalBalance       int64
	Accounts           []*ports.AccountBalance
	Jars               []*ports.JarSummary
	Goals              []*ports.GoalSummary
	RecentTransactions []*ports.TransactionListItem
}

// GetDashboard retrieves and composes the complete dashboard
func (s *DashboardService) GetDashboard(ctx context.Context) (*DashboardResponse, error) {
	s.logger.DebugContext(ctx, "fetching dashboard data")

	totalBalance, err := s.dashboardRepo.GetTotalBalance(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get total balance", slog.String("error", err.Error()))
		return nil, err
	}

	summary, err := s.dashboardRepo.GetMonthlySummary(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get monthly summary", slog.String("error", err.Error()))
		return nil, err
	}
	summary.TotalBalance = totalBalance

	accounts, err := s.dashboardRepo.GetAccountBalances(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get account balances", slog.String("error", err.Error()))
		return nil, err
	}

	jars, err := s.dashboardRepo.GetJarSummaries(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get jar summaries", slog.String("error", err.Error()))
		return nil, err
	}

	goals, err := s.dashboardRepo.GetGoalSummaries(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get goal summaries", slog.String("error", err.Error()))
		return nil, err
	}

	recentTx, err := s.dashboardRepo.GetRecentTransactions(ctx, 5)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get recent transactions", slog.String("error", err.Error()))
		return nil, err
	}

	dashboard := &DashboardResponse{
		Summary:            summary,
		TotalBalance:       totalBalance,
		Accounts:           accounts,
		Jars:               jars,
		Goals:              goals,
		RecentTransactions: recentTx,
	}

	s.logger.InfoContext(ctx, "dashboard data fetched successfully")
	return dashboard, nil
}
