package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
)

type DashboardSummaryResponse struct {
	TotalBalance       int64   `json:"total_balance"`
	MonthlyIncome      int64   `json:"monthly_income"`
	MonthlyExpense     int64   `json:"monthly_expense"`
	MonthlySavings     int64   `json:"monthly_savings"`
	MonthlySavingsRate float64 `json:"monthly_savings_rate"`
}

type DashboardAccountResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Balance int64     `json:"balance"`
}

type DashboardJarResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	AllocationType      string    `json:"allocation_type"`
	AllocationValue     int64     `json:"allocation_value"`
	Allocated           int64     `json:"allocated"`
	Used                int64     `json:"used"`
	Available           int64     `json:"available"`
	UsedPercentage      float64   `json:"used_percentage"`
	AvailablePercentage float64   `json:"available_percentage"`
}

type DashboardGoalResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Target      int64     `json:"target"`
	Contributed int64     `json:"contributed"`
	Remaining   int64     `json:"remaining"`
	Progress    float64   `json:"progress"`
	Deadline    time.Time `json:"deadline"`
}

type DashboardResponse struct {
	Summary            *DashboardSummaryResponse      `json:"summary"`
	Accounts           []*DashboardAccountResponse    `json:"accounts"`
	Jars               []*DashboardJarResponse        `json:"jars"`
	Goals              []*DashboardGoalResponse       `json:"goals"`
	RecentTransactions []*TransactionListItemResponse `json:"recent_transactions"`
}

// NewDashboardResponse converts domain types to HTTP response
func NewDashboardResponse(dash *application.DashboardResponse) *DashboardResponse {
	// Convert summary
	summaryResp := &DashboardSummaryResponse{
		TotalBalance:       dash.TotalBalance,
		MonthlyIncome:      dash.Summary.MonthlyIncome,
		MonthlyExpense:     dash.Summary.MonthlyExpense,
		MonthlySavings:     dash.Summary.MonthlySavings,
		MonthlySavingsRate: dash.Summary.MonthlySavingsRate,
	}

	// Convert accounts
	accountsResp := make([]*DashboardAccountResponse, len(dash.Accounts))
	for i, a := range dash.Accounts {
		accountsResp[i] = &DashboardAccountResponse{
			ID:      a.ID,
			Name:    a.Name,
			Balance: a.Balance,
		}
	}

	// Convert jars
	jarsResp := make([]*DashboardJarResponse, len(dash.Jars))
	for i, j := range dash.Jars {
		jarsResp[i] = &DashboardJarResponse{
			ID:                  j.ID,
			Name:                j.Name,
			AllocationType:      j.AllocationType,
			AllocationValue:     j.AllocationValue,
			Allocated:           j.Allocated,
			Used:                j.Used,
			Available:           j.Available,
			UsedPercentage:      j.UsedPercentage,
			AvailablePercentage: j.AvailablePercentage,
		}
	}

	// Convert goals
	goalsResp := make([]*DashboardGoalResponse, len(dash.Goals))
	for i, g := range dash.Goals {
		goalsResp[i] = &DashboardGoalResponse{
			ID:          g.ID,
			Name:        g.Name,
			Target:      g.Target,
			Contributed: g.Contributed,
			Remaining:   g.Remaining,
			Progress:    g.Progress,
			Deadline:    g.Deadline,
		}
	}

	// Convert recent transactions
	txResp := make([]*TransactionListItemResponse, len(dash.RecentTransactions))
	for i, t := range dash.RecentTransactions {
		resp := transactionListItemToResponse(t)
		txResp[i] = &resp
	}

	return &DashboardResponse{
		Summary:            summaryResp,
		Accounts:           accountsResp,
		Jars:               jarsResp,
		Goals:              goalsResp,
		RecentTransactions: txResp,
	}
}
