package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
	"github.com/stretchr/testify/require"
)

type mockStatementService struct {
	getFn func(context.Context, uuid.UUID, ports.ListParams) (*application.Statement, error)
}

func (m *mockStatementService) Get(
	ctx context.Context,
	accountID uuid.UUID,
	params ports.ListParams,
) (*application.Statement, error) {
	return m.getFn(ctx, accountID, params)
}

func doStatementRequest(
	t *testing.T,
	svc *mockStatementService,
	accountID uuid.UUID,
	query string,
) *httptest.ResponseRecorder {
	t.Helper()

	h := handler.NewStatementHandler(svc, newTestLogger())
	req := httptest.NewRequest(
		http.MethodGet,
		"/accounts/"+accountID.String()+"/statement"+query,
		nil,
	)
	req = withAccountID(req, accountID)
	rec := httptest.NewRecorder()

	h.HandleGet(rec, req)

	return rec
}

func TestStatementHandler_ReturnsAccountSummaryRowsAndPagination(t *testing.T) {
	accountID := uuid.New()
	itemID := uuid.New()
	jarName := "Needs"
	balance := int64(1122000)

	svc := &mockStatementService{
		getFn: func(
			_ context.Context,
			gotID uuid.UUID,
			params ports.ListParams,
		) (*application.Statement, error) {
			require.Equal(t, accountID, gotID)
			require.Equal(t, 20, params.Limit)
			require.Equal(t, 0, params.Offset)

			return &application.Statement{
				Account: &account.Account{
					ID:        accountID,
					Name:      "SBI",
					IconKey:   "sbi",
					IsPrimary: true,
					Balance:   1122000,
				},
				Summary: ports.StatementSummary{
					OpeningBalance: 0,
					ClosingBalance: 1122000,
					Inflow:         1200000,
					Outflow:        78000,
				},
				Transactions: []*ports.TransactionListItem{{
					ID:             itemID,
					Name:           "Salary",
					Type:           transaction.TransactionTypeIncome,
					Amount:         1200000,
					JarName:        &jarName,
					AccountBalance: &balance,
					Category:       transaction.TransactionCategoryOther,
					OccurredAt:     time.Date(2026, 10, 1, 5, 12, 31, 0, time.UTC),
				}},
				Total: 2,
			}, nil
		},
	}

	rec := doStatementRequest(t, svc, accountID, "")
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Account      map[string]any   `json:"account"`
		Summary      map[string]any   `json:"summary"`
		Transactions []map[string]any `json:"transactions"`
		Pagination   map[string]any   `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	require.Equal(t, "SBI", body.Account["name"])
	require.Equal(t, "sbi", body.Account["icon_key"])
	require.Equal(t, true, body.Account["is_primary"])
	require.Equal(t, float64(1122000), body.Account["balance"])

	require.Equal(t, float64(0), body.Summary["opening_balance"])
	require.Equal(t, float64(1122000), body.Summary["closing_balance"])
	require.Equal(t, float64(1200000), body.Summary["inflow"])
	require.Equal(t, float64(78000), body.Summary["outflow"])
	require.Equal(t, float64(1122000), body.Summary["net"])

	require.Len(t, body.Transactions, 1)
	row := body.Transactions[0]
	require.Equal(t, itemID.String(), row["id"])
	require.Equal(t, "Salary", row["name"])
	require.Equal(t, "2026-10-01T10:42:31+05:30", row["occurred_at"])
	require.Equal(t, "Needs", row["jar_name"])
	require.Equal(t, "other", row["category"])
	require.Equal(t, float64(1200000), row["amount"])
	require.Equal(t, float64(1122000), row["balance_after"])

	// 1 row returned out of 2 matching: there is a next page.
	require.Equal(t, float64(20), body.Pagination["limit"])
	require.Equal(t, float64(0), body.Pagination["offset"])
	require.Equal(t, float64(2), body.Pagination["total"])
	require.Equal(t, true, body.Pagination["has_next"])
}

func TestStatementHandler_ParsesFilters(t *testing.T) {
	accountID := uuid.New()
	jarID := uuid.New()

	var got ports.ListParams

	svc := &mockStatementService{
		getFn: func(
			_ context.Context,
			_ uuid.UUID,
			params ports.ListParams,
		) (*application.Statement, error) {
			got = params

			return &application.Statement{
				Account: &account.Account{ID: accountID},
			}, nil
		},
	}

	// A single day: same plain date for both bounds.
	query := fmt.Sprintf(
		"?from_date=2026-09-09&to_date=2026-09-09&type=expense"+
			"&category=entertainment&jar_id=%s&search=movie&limit=50&offset=100",
		jarID,
	)

	rec := doStatementRequest(t, svc, accountID, query)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 50, got.Limit)
	require.Equal(t, 100, got.Offset)
	require.Equal(t, "movie", *got.Search)
	require.Equal(t, transaction.TransactionTypeExpense, *got.Type)
	require.Equal(t, transaction.TransactionCategoryEntertainment, *got.Category)
	require.Equal(t, jarID, *got.JarID)

	// Single-day filter must cover exactly one IST day.
	wantFrom := time.Date(2026, 9, 9, 0, 0, 0, 0, timeutil.IST).UTC()
	wantTo := time.Date(2026, 9, 10, 0, 0, 0, 0, timeutil.IST).UTC()
	require.True(t, got.FromDate.Equal(wantFrom), "from = %v", got.FromDate)
	require.True(t, got.ToDate.Equal(wantTo), "to = %v", got.ToDate)
}

func TestStatementHandler_HasNextFalseOnLastPage(t *testing.T) {
	accountID := uuid.New()

	svc := &mockStatementService{
		getFn: func(
			context.Context,
			uuid.UUID,
			ports.ListParams,
		) (*application.Statement, error) {
			return &application.Statement{
				Account: &account.Account{ID: accountID},
				Transactions: []*ports.TransactionListItem{
					{ID: uuid.New(), OccurredAt: time.Now()},
				},
				Total: 21,
			}, nil
		},
	}

	rec := doStatementRequest(t, svc, accountID, "?limit=20&offset=20")
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Pagination struct {
			HasNext bool `json:"has_next"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.False(t, body.Pagination.HasNext)
}

func TestStatementHandler_RejectsInvalidParams(t *testing.T) {
	cases := map[string]string{
		"INVALID_LIMIT":      "?limit=0",
		"INVALID_LIMIT_MAX":  "?limit=101",
		"INVALID_OFFSET":     "?offset=-1",
		"INVALID_TYPE":       "?type=bogus",
		"INVALID_CATEGORY":   "?category=bogus",
		"INVALID_JAR_ID":     "?jar_id=not-a-uuid",
		"INVALID_FROM_DATE":  "?from_date=yesterday",
		"INVALID_TO_DATE":    "?to_date=31-12-2026",
		"INVALID_DATE_RANGE": "?from_date=2026-10-02&to_date=2026-10-01",
	}

	for name, query := range cases {
		t.Run(name, func(t *testing.T) {
			svc := &mockStatementService{
				getFn: func(
					context.Context,
					uuid.UUID,
					ports.ListParams,
				) (*application.Statement, error) {
					t.Fatal("service must not be called for invalid params")
					return nil, nil
				},
			}

			rec := doStatementRequest(t, svc, uuid.New(), query)
			require.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestStatementHandler_AccountNotFound(t *testing.T) {
	svc := &mockStatementService{
		getFn: func(
			context.Context,
			uuid.UUID,
			ports.ListParams,
		) (*application.Statement, error) {
			return nil, fmt.Errorf("get statement: %w", account.ErrAccountNotFound)
		},
	}

	rec := doStatementRequest(t, svc, uuid.New(), "")
	require.Equal(t, http.StatusNotFound, rec.Code)
}
