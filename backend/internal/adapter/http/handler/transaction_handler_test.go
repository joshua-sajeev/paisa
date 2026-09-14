package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/stretchr/testify/require"
)

type mockTransactionService struct {
	listFn          func(context.Context, ports.ListParams) ([]*ports.TransactionListItem, error)
	listByAccountFn func(context.Context, uuid.UUID, ports.ListParams) ([]*ports.TransactionListItem, error)
}

func (m *mockTransactionService) Create(
	context.Context,
	string,
	transaction.TransactionType,
	transaction.TransactionCategory,
	*uuid.UUID,
	*uuid.UUID,
	*uuid.UUID,
	int64,
	time.Time,
	bool,
) (*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionService) Update(
	context.Context,
	uuid.UUID,
	string,
	transaction.TransactionCategory,
	*uuid.UUID,
	*uuid.UUID,
	*uuid.UUID,
	int64,
	time.Time,
	bool,
) (*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockTransactionService) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (m *mockTransactionService) List(
	ctx context.Context,
	params ports.ListParams,
) ([]*ports.TransactionListItem, error) {
	return m.listFn(ctx, params)
}

func (m *mockTransactionService) ListByAccount(
	ctx context.Context,
	accountID uuid.UUID,
	params ports.ListParams,
) ([]*ports.TransactionListItem, error) {
	return m.listByAccountFn(ctx, accountID, params)
}

func (m *mockTransactionService) GetByID(
	context.Context,
	uuid.UUID,
) (*transaction.Transaction, error) {
	return nil, nil
}

func TestTransactionHandler_HandleListReturnsDisplayFields(t *testing.T) {
	jarName := "Needs"
	itemID := uuid.New()

	service := &mockTransactionService{
		listFn: func(
			_ context.Context,
			params ports.ListParams,
		) ([]*ports.TransactionListItem, error) {
			require.Equal(t, 10, params.Limit)
			require.Equal(t, 5, params.Offset)

			return []*ports.TransactionListItem{
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
			}, nil
		},
	}

	h := handler.NewTransactionHandler(service, newTestLogger())
	req := httptest.NewRequest(
		http.MethodGet,
		"/transactions?limit=10&offset=5",
		nil,
	)
	rec := httptest.NewRecorder()

	h.HandleList(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Transactions []map[string]any `json:"transactions"`
		Total        int              `json:"total"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 1, body.Total)
	require.Len(t, body.Transactions, 1)

	got := body.Transactions[0]
	require.Equal(t, itemID.String(), got["id"])
	require.Equal(t, "Move to savings", got["name"])
	require.Equal(t, "2026-01-03", got["date"])
	require.Equal(t, "transfer", got["type"])
	require.Equal(t, float64(300), got["amount"])
	require.Equal(t, "Needs", got["jar_name"])
	require.Equal(t, "Checking -> Savings", got["account"])
	require.Nil(t, got["account_balance"])
	require.Equal(t, "transfer", got["category"])
	require.NotContains(t, got, "occurred_at")
	require.NotContains(t, got, "balance_after")
}

func TestTransactionHandler_HandleListByAccountReturnsAccountProjection(t *testing.T) {
	accountID := uuid.New()
	itemID := uuid.New()
	balance := int64(500)

	service := &mockTransactionService{
		listByAccountFn: func(
			_ context.Context,
			gotAccountID uuid.UUID,
			params ports.ListParams,
		) ([]*ports.TransactionListItem, error) {
			require.Equal(t, accountID, gotAccountID)
			require.Equal(t, 20, params.Limit)
			require.Equal(t, 0, params.Offset)

			return []*ports.TransactionListItem{
				{
					ID:             itemID,
					Name:           "Move to savings",
					Type:           transaction.TransactionTypeTransfer,
					Amount:         -300,
					Account:        "Savings",
					AccountBalance: &balance,
					Category:       transaction.TransactionCategoryTransfer,
					OccurredAt:     time.Date(2026, 1, 3, 9, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	h := handler.NewTransactionHandler(service, newTestLogger())
	req := httptest.NewRequest(http.MethodGet, "/accounts/"+accountID.String()+"/transactions", nil)
	req = withAccountID(req, accountID)
	rec := httptest.NewRecorder()

	h.HandleListByAccount(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Transactions []map[string]any `json:"transactions"`
		Total        int              `json:"total"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 1, body.Total)
	require.Len(t, body.Transactions, 1)

	got := body.Transactions[0]
	require.Equal(t, itemID.String(), got["id"])
	require.Equal(t, "2026-01-03", got["date"])
	require.Equal(t, "Savings", got["account"])
	require.Equal(t, float64(-300), got["amount"])
	require.Equal(t, float64(500), got["account_balance"])
	require.Equal(t, "transfer", got["category"])
	require.NotContains(t, got, "occurred_at")
	require.NotContains(t, got, "balance_after")
}
