package application_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
	"github.com/stretchr/testify/require"
)

type mockAllocationRepository struct {
	createFn                func(context.Context, *transaction.Allocation) error
	deleteByTransactionIDFn func(context.Context, uuid.UUID) error
	listFn                  func(context.Context, ports.AllocationListParams) ([]*ports.AllocationListItem, int, error)
	getMonthlySummaryFn     func(context.Context, time.Time, time.Time, *uuid.UUID) (int64, error)
}

func (m *mockAllocationRepository) Create(ctx context.Context, alloc *transaction.Allocation) error {
	if m.createFn != nil {
		return m.createFn(ctx, alloc)
	}
	return nil
}

func (m *mockAllocationRepository) DeleteByTransactionID(ctx context.Context, txnID uuid.UUID) error {
	if m.deleteByTransactionIDFn != nil {
		return m.deleteByTransactionIDFn(ctx, txnID)
	}
	return nil
}

func (m *mockAllocationRepository) List(
	ctx context.Context,
	params ports.AllocationListParams,
) ([]*ports.AllocationListItem, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockAllocationRepository) GetMonthlySummary(
	ctx context.Context,
	start time.Time,
	end time.Time,
	jarID *uuid.UUID,
) (int64, error) {
	if m.getMonthlySummaryFn != nil {
		return m.getMonthlySummaryFn(ctx, start, end, jarID)
	}
	return 0, nil
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestAllocationService_List_SuccessWithExplicitMonth(t *testing.T) {
	ctx := context.Background()
	jarID := uuid.New()
	monthParam := "2026-08"

	expectedStart := time.Date(2026, 8, 1, 0, 0, 0, 0, timeutil.IST).UTC()
	expectedEnd := time.Date(2026, 9, 1, 0, 0, 0, 0, timeutil.IST).UTC()

	repo := &mockAllocationRepository{
		listFn: func(
			_ context.Context,
			params ports.AllocationListParams,
		) ([]*ports.AllocationListItem, int, error) {
			require.Equal(t, 15, params.Limit)
			return []*ports.AllocationListItem{
				{
					ID:              uuid.New(),
					TransactionID:   uuid.New(),
					TransactionName: "August Salary",
					JarID:           jarID,
					JarName:         "Investments",
					AllocationType:  jar.AllocationTypePercentage,
					Amount:          200000,
					OccurredAt:      expectedStart,
					CreatedAt:       expectedStart,
				},
			}, 1, nil
		},
		getMonthlySummaryFn: func(
			_ context.Context,
			start time.Time,
			end time.Time,
			jID *uuid.UUID,
		) (int64, error) {
			require.True(t, expectedStart.Equal(start))
			require.True(t, expectedEnd.Equal(end))

			if jID == nil {
				return 500000, nil
			}
			require.Equal(t, jarID, *jID)
			return 200000, nil
		},
	}

	svc := application.NewAllocationService(repo, newDiscardLogger())

	result, err := svc.List(ctx, ports.AllocationListParams{
		Limit: 15,
		JarID: &jarID,
	}, &monthParam)

	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
	require.Len(t, result.Allocations, 1)
	require.Equal(t, "August Salary", result.Allocations[0].TransactionName)
	require.Equal(t, "2026-08", result.MonthlySummary.Month)
	require.Equal(t, int64(500000), result.MonthlySummary.TotalAllocated)
	require.NotNil(t, result.MonthlySummary.JarTotalAllocated)
	require.Equal(t, int64(200000), *result.MonthlySummary.JarTotalAllocated)
}

func TestAllocationService_List_FallbackToFromDate(t *testing.T) {
	ctx := context.Background()
	fromDate := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	expectedStart := time.Date(2026, 4, 1, 0, 0, 0, 0, timeutil.IST).UTC()
	expectedEnd := time.Date(2026, 5, 1, 0, 0, 0, 0, timeutil.IST).UTC()

	repo := &mockAllocationRepository{
		listFn: func(
			_ context.Context,
			_ ports.AllocationListParams,
		) ([]*ports.AllocationListItem, int, error) {
			return nil, 0, nil
		},
		getMonthlySummaryFn: func(
			_ context.Context,
			start time.Time,
			end time.Time,
			_ *uuid.UUID,
		) (int64, error) {
			require.True(t, expectedStart.Equal(start))
			require.True(t, expectedEnd.Equal(end))
			return 0, nil
		},
	}

	svc := application.NewAllocationService(repo, newDiscardLogger())

	result, err := svc.List(ctx, ports.AllocationListParams{
		FromDate: &fromDate,
	}, nil)

	require.NoError(t, err)
	require.Equal(t, "2026-04", result.MonthlySummary.Month)
}

func TestAllocationService_List_FallbackToCurrentMonth(t *testing.T) {
	ctx := context.Background()
	nowIST := time.Now().In(timeutil.IST)
	expectedMonth := nowIST.Format("2006-01")

	repo := &mockAllocationRepository{
		listFn: func(
			_ context.Context,
			_ ports.AllocationListParams,
		) ([]*ports.AllocationListItem, int, error) {
			return nil, 0, nil
		},
		getMonthlySummaryFn: func(
			_ context.Context,
			_ time.Time,
			_ time.Time,
			_ *uuid.UUID,
		) (int64, error) {
			return 1000, nil
		},
	}

	svc := application.NewAllocationService(repo, newDiscardLogger())

	result, err := svc.List(ctx, ports.AllocationListParams{}, nil)

	require.NoError(t, err)
	require.Equal(t, expectedMonth, result.MonthlySummary.Month)
	require.Equal(t, int64(1000), result.MonthlySummary.TotalAllocated)
	require.Nil(t, result.MonthlySummary.JarTotalAllocated)
}

func TestAllocationService_List_InvalidMonthFormat(t *testing.T) {
	ctx := context.Background()
	invalidMonth := "bad-format"
	repo := &mockAllocationRepository{}
	svc := application.NewAllocationService(repo, newDiscardLogger())

	_, err := svc.List(ctx, ports.AllocationListParams{}, &invalidMonth)
	require.ErrorIs(t, err, application.ErrInvalidMonthFormat)
}

func TestAllocationService_List_RepoErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("list error", func(t *testing.T) {
		repo := &mockAllocationRepository{
			listFn: func(
				_ context.Context,
				_ ports.AllocationListParams,
			) ([]*ports.AllocationListItem, int, error) {
				return nil, 0, errors.New("query failed")
			},
		}
		svc := application.NewAllocationService(repo, newDiscardLogger())
		_, err := svc.List(ctx, ports.AllocationListParams{}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "list allocations")
	})

	t.Run("monthly summary error", func(t *testing.T) {
		repo := &mockAllocationRepository{
			listFn: func(
				_ context.Context,
				_ ports.AllocationListParams,
			) ([]*ports.AllocationListItem, int, error) {
				return nil, 0, nil
			},
			getMonthlySummaryFn: func(
				_ context.Context,
				_ time.Time,
				_ time.Time,
				_ *uuid.UUID,
			) (int64, error) {
				return 0, errors.New("summary failed")
			},
		}
		svc := application.NewAllocationService(repo, newDiscardLogger())
		_, err := svc.List(ctx, ports.AllocationListParams{}, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "get monthly allocation summary")
	})
}
