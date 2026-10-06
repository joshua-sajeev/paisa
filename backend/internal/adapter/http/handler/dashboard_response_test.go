package handler_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/stretchr/testify/require"
)

func TestNewDashboardResponseMapsRecentTransactionsDisplayFields(t *testing.T) {
	jarName := "Needs"
	itemID := uuid.New()

	dashboard := &application.DashboardResponse{
		Summary:      &ports.DashboardSummary{},
		TotalBalance: 1250,
		Accounts:     []*ports.AccountBalance{},
		Jars:         []*ports.JarSummary{},
		Goals:        []*ports.GoalSummary{},
		RecentTransactions: []*ports.TransactionListItem{
			{
				ID:             itemID,
				Name:           "Move to savings",
				Type:           transaction.TransactionTypeTransfer,
				Amount:         300,
				JarName:        &jarName,
				Account:        "Checking -> Savings",
				AccountBalance: nil,
				Category:       transaction.TransactionCategoryTransfer,
				OccurredAt:     time.Date(2026, 1, 3, 9, 0, 0, 0, time.UTC),
			},
		},
	}

	resp := handler.NewDashboardResponse(dashboard)
	bodyBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(bodyBytes, &body))

	recent, ok := body["recent_transactions"].([]any)
	require.True(t, ok)
	require.Len(t, recent, 1)

	got, ok := recent[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, itemID.String(), got["id"])
	require.Equal(t, "Move to savings", got["name"])
	require.Equal(t, "2026-01-03T14:30:00+05:30", got["occurred_at"])
	require.NotContains(t, got, "date")
	require.Equal(t, "transfer", got["type"])
	require.Equal(t, float64(300), got["amount"])
	require.Equal(t, "Needs", got["jar_name"])
	require.Equal(t, "Checking -> Savings", got["account"])
	require.Nil(t, got["account_balance"])
	require.Equal(t, "transfer", got["category"])
	require.NotContains(t, got, "balance_after")
}

func TestNewDashboardResponseMapsJarIconKey(t *testing.T) {
	jarID := uuid.New()

	dashboard := &application.DashboardResponse{
		Summary:      &ports.DashboardSummary{},
		TotalBalance: 0,
		Accounts:     []*ports.AccountBalance{},
		Jars: []*ports.JarSummary{
			{ID: jarID, Name: "Needs", IconKey: "piggy"},
		},
		Goals:              []*ports.GoalSummary{},
		RecentTransactions: []*ports.TransactionListItem{},
	}

	resp := handler.NewDashboardResponse(dashboard)
	bodyBytes, err := json.Marshal(resp)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(bodyBytes, &body))

	jars, ok := body["jars"].([]any)
	require.True(t, ok)
	require.Len(t, jars, 1)

	got, ok := jars[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, jarID.String(), got["id"])
	require.Equal(t, "piggy", got["icon_key"])
}
