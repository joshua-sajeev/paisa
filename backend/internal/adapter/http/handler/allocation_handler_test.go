package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
	"github.com/stretchr/testify/require"
)

type mockAllocationService struct {
	listFn func(
		ctx context.Context,
		params ports.AllocationListParams,
		monthParam *string,
	) (*application.AllocationListResult, error)
}

func (m *mockAllocationService) List(
	ctx context.Context,
	params ports.AllocationListParams,
	monthParam *string,
) (*application.AllocationListResult, error) {
	return m.listFn(ctx, params, monthParam)
}

func TestAllocationHandler_HandleList_Success(t *testing.T) {
	allocID := uuid.New()
	txnID := uuid.New()
	jarID := uuid.New()
	occurredAt := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 10, 2, 12, 1, 0, 0, time.UTC)
	jarTotal := int64(150000)

	service := &mockAllocationService{
		listFn: func(
			_ context.Context,
			params ports.AllocationListParams,
			monthParam *string,
		) (*application.AllocationListResult, error) {
			require.Equal(t, 10, params.Limit)
			require.Equal(t, 20, params.Offset)
			require.NotNil(t, params.Search)
			require.Equal(t, "salary", *params.Search)
			require.NotNil(t, params.JarID)
			require.Equal(t, jarID, *params.JarID)
			require.NotNil(t, params.AllocationType)
			require.Equal(t, jar.AllocationTypePercentage, *params.AllocationType)
			require.NotNil(t, monthParam)
			require.Equal(t, "2026-10", *monthParam)

			return &application.AllocationListResult{
				Allocations: []*ports.AllocationListItem{
					{
						ID:              allocID,
						TransactionID:   txnID,
						TransactionName: "October Salary",
						JarID:           jarID,
						JarName:         "Necessities",
						AllocationType:  jar.AllocationTypePercentage,
						Amount:          150000,
						OccurredAt:      occurredAt,
						CreatedAt:       createdAt,
					},
				},
				Total: 1,
				MonthlySummary: ports.AllocationMonthlySummary{
					Month:             "2026-10",
					TotalAllocated:    300000,
					JarTotalAllocated: &jarTotal,
				},
			}, nil
		},
	}

	h := handler.NewAllocationHandler(service, newTestLogger())

	req := httptest.NewRequest(
		http.MethodGet,
		"/allocations?limit=10&offset=20&search=salary&jar_id="+jarID.String()+"&allocation_type=percentage&month=2026-10",
		nil,
	)
	rec := httptest.NewRecorder()

	h.HandleList(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handler.ListAllocationsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	require.Equal(t, 1, resp.Total)
	require.Equal(t, 10, resp.Limit)
	require.Equal(t, 20, resp.Offset)
	require.Len(t, resp.Allocations, 1)

	item := resp.Allocations[0]
	require.Equal(t, allocID.String(), item.ID)
	require.Equal(t, txnID.String(), item.TransactionID)
	require.Equal(t, "October Salary", item.TransactionName)
	require.Equal(t, jarID.String(), item.JarID)
	require.Equal(t, "Necessities", item.JarName)
	require.Equal(t, "percentage", item.AllocationType)
	require.Equal(t, int64(150000), item.Amount)
	require.True(t, occurredAt.In(timeutil.IST).Equal(item.OccurredAt))
	require.True(t, createdAt.Equal(item.CreatedAt))

	require.Equal(t, "2026-10", resp.MonthlySummary.Month)
	require.Equal(t, int64(300000), resp.MonthlySummary.TotalAllocated)
	require.NotNil(t, resp.MonthlySummary.JarTotalAllocated)
	require.Equal(t, int64(150000), *resp.MonthlySummary.JarTotalAllocated)
}

func TestAllocationHandler_HandleList_InvalidJarID(t *testing.T) {
	service := &mockAllocationService{}
	h := handler.NewAllocationHandler(service, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/allocations?jar_id=invalid-uuid", nil)
	rec := httptest.NewRecorder()

	h.HandleList(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAllocationHandler_HandleList_InvalidAllocationType(t *testing.T) {
	service := &mockAllocationService{}
	h := handler.NewAllocationHandler(service, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/allocations?allocation_type=invalid_type", nil)
	rec := httptest.NewRecorder()

	h.HandleList(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAllocationHandler_HandleList_InvalidMonth(t *testing.T) {
	service := &mockAllocationService{
		listFn: func(
			_ context.Context,
			_ ports.AllocationListParams,
			_ *string,
		) (*application.AllocationListResult, error) {
			return nil, application.ErrInvalidMonthFormat
		},
	}
	h := handler.NewAllocationHandler(service, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/allocations?month=2026-13", nil)
	rec := httptest.NewRecorder()

	h.HandleList(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAllocationHandler_HandleList_InternalError(t *testing.T) {
	service := &mockAllocationService{
		listFn: func(
			_ context.Context,
			_ ports.AllocationListParams,
			_ *string,
		) (*application.AllocationListResult, error) {
			return nil, errors.New("db error")
		},
	}
	h := handler.NewAllocationHandler(service, newTestLogger())

	req := httptest.NewRequest(http.MethodGet, "/allocations", nil)
	rec := httptest.NewRecorder()

	h.HandleList(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
