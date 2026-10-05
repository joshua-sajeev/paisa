package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
)

var ErrInvalidMonthFormat = errors.New("invalid month format, expected YYYY-MM")

// AllocationListResult is the output of AllocationService.List.
type AllocationListResult struct {
	Allocations    []*ports.AllocationListItem
	Total          int
	MonthlySummary ports.AllocationMonthlySummary
}

// AllocationService coordinates jar allocation read queries.
type AllocationService struct {
	allocationRepo ports.AllocationRepository
	logger         *slog.Logger
}

// NewAllocationService creates a new AllocationService.
func NewAllocationService(
	allocationRepo ports.AllocationRepository,
	logger *slog.Logger,
) *AllocationService {
	return &AllocationService{
		allocationRepo: allocationRepo,
		logger:         logger,
	}
}

// List returns a paginated list of allocations matching filters, along with a monthly summary.
func (s *AllocationService) List(
	ctx context.Context,
	params ports.AllocationListParams,
	monthParam *string,
) (*AllocationListResult, error) {
	targetMonth, err := s.resolveMonth(params, monthParam)
	if err != nil {
		return nil, err
	}

	monthStart := targetMonth.UTC()
	monthEnd := targetMonth.AddDate(0, 1, 0).UTC()
	monthStr := targetMonth.Format("2006-01")

	allocations, total, err := s.allocationRepo.List(ctx, params)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to list allocations",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("list allocations: %w", err)
	}

	totalAllocated, err := s.allocationRepo.GetMonthlySummary(ctx, monthStart, monthEnd, nil)
	if err != nil {
		s.logger.ErrorContext(
			ctx,
			"failed to get monthly allocation summary",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get monthly allocation summary: %w", err)
	}

	summary := ports.AllocationMonthlySummary{
		Month:          monthStr,
		TotalAllocated: totalAllocated,
	}

	if params.JarID != nil {
		jarTotal, err := s.allocationRepo.GetMonthlySummary(ctx, monthStart, monthEnd, params.JarID)
		if err != nil {
			s.logger.ErrorContext(
				ctx,
				"failed to get monthly jar allocation summary",
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("get monthly jar allocation summary: %w", err)
		}
		summary.JarTotalAllocated = &jarTotal
	}

	return &AllocationListResult{
		Allocations:    allocations,
		Total:          total,
		MonthlySummary: summary,
	}, nil
}

func (s *AllocationService) resolveMonth(
	params ports.AllocationListParams,
	monthParam *string,
) (time.Time, error) {
	if monthParam != nil && *monthParam != "" {
		parsed, err := time.ParseInLocation("2006-01", *monthParam, timeutil.IST)
		if err != nil {
			return time.Time{}, ErrInvalidMonthFormat
		}
		return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, timeutil.IST), nil
	}

	if params.FromDate != nil && params.ToDate != nil {
		fromIST := params.FromDate.In(timeutil.IST)
		// ToDate is exclusive. Check month of last inclusive moment.
		toInclusiveIST := params.ToDate.Add(-1 * time.Nanosecond).In(timeutil.IST)

		if fromIST.Year() == toInclusiveIST.Year() && fromIST.Month() == toInclusiveIST.Month() {
			return time.Date(fromIST.Year(), fromIST.Month(), 1, 0, 0, 0, 0, timeutil.IST), nil
		}
	}

	nowIST := time.Now().In(timeutil.IST)
	return time.Date(nowIST.Year(), nowIST.Month(), 1, 0, 0, 0, 0, timeutil.IST), nil
}
