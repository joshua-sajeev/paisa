package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
)

// AllocationService defines the interface required by AllocationHandler.
type AllocationService interface {
	List(
		ctx context.Context,
		params ports.AllocationListParams,
		monthParam *string,
	) (*application.AllocationListResult, error)
}

// AllocationHandler handles HTTP requests for jar allocations.
type AllocationHandler struct {
	service AllocationService
	logger  *slog.Logger
}

// NewAllocationHandler creates a new AllocationHandler.
func NewAllocationHandler(
	service AllocationService,
	logger *slog.Logger,
) *AllocationHandler {
	return &AllocationHandler{
		service: service,
		logger:  logger,
	}
}

// AllocationListItemResponse represents an allocation in API list responses.
type AllocationListItemResponse struct {
	ID              string    `json:"id"`
	TransactionID   string    `json:"transaction_id"`
	TransactionName string    `json:"transaction_name"`
	JarID           string    `json:"jar_id"`
	JarName         string    `json:"jar_name"`
	AllocationType  string    `json:"allocation_type"`
	Amount          int64     `json:"amount"`
	OccurredAt      time.Time `json:"occurred_at"`
	CreatedAt       time.Time `json:"created_at"`
}

func allocationListItemToResponse(
	item *ports.AllocationListItem,
) AllocationListItemResponse {
	return AllocationListItemResponse{
		ID:              item.ID.String(),
		TransactionID:   item.TransactionID.String(),
		TransactionName: item.TransactionName,
		JarID:           item.JarID.String(),
		JarName:         item.JarName,
		AllocationType:  string(item.AllocationType),
		Amount:          item.Amount,
		OccurredAt:      item.OccurredAt.In(timeutil.IST),
		CreatedAt:       item.CreatedAt.UTC(),
	}
}

// AllocationMonthlySummaryResponse represents the monthly summary in response.
type AllocationMonthlySummaryResponse struct {
	Month             string `json:"month"`
	TotalAllocated    int64  `json:"total_allocated"`
	JarTotalAllocated *int64 `json:"jar_total_allocated,omitempty"`
}

// ListAllocationsResponse represents the global allocation list response.
type ListAllocationsResponse struct {
	Allocations    []AllocationListItemResponse     `json:"allocations"`
	Total          int                              `json:"total"`
	Limit          int                              `json:"limit"`
	Offset         int                              `json:"offset"`
	MonthlySummary AllocationMonthlySummaryResponse `json:"monthly_summary"`
}

// HandleList handles GET /allocations.
func (h *AllocationHandler) HandleList(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	params := ports.AllocationListParams{
		Limit:  limit,
		Offset: offset,
	}

	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = &search
	}

	if jarIDStr := r.URL.Query().Get("jar_id"); jarIDStr != "" {
		jarID, err := uuid.Parse(jarIDStr)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid jar_id", "ERR_INVALID_ID")
			return
		}
		params.JarID = &jarID
	}

	if jarName := r.URL.Query().Get("jar_name"); jarName != "" {
		params.JarName = &jarName
	}

	if txnName := r.URL.Query().Get("transaction_name"); txnName != "" {
		params.TransactionName = &txnName
	}

	if typeStr := r.URL.Query().Get("allocation_type"); typeStr != "" {
		allocType := jar.AllocationType(typeStr)
		if !allocType.IsValid() {
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid allocation_type", "ERR_INVALID_ALLOCATION_TYPE")
			return
		}
		params.AllocationType = &allocType
	}

	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		if fromDate, ok := parseFilterTime(fromDateStr, false); ok {
			params.FromDate = &fromDate
		}
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		if toDate, ok := parseFilterTime(toDateStr, true); ok {
			params.ToDate = &toDate
		}
	}

	if minAmountStr := r.URL.Query().Get("min_amount"); minAmountStr != "" {
		if minAmount, err := strconv.ParseInt(minAmountStr, 10, 64); err == nil {
			params.MinAmount = &minAmount
		}
	}

	if maxAmountStr := r.URL.Query().Get("max_amount"); maxAmountStr != "" {
		if maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64); err == nil {
			params.MaxAmount = &maxAmount
		}
	}

	var monthParam *string
	if m := r.URL.Query().Get("month"); m != "" {
		monthParam = &m
	} else if params.FromDate != nil && params.ToDate != nil {
		fromIST := params.FromDate.In(timeutil.IST)
		// toInclusiveIST is the last moment of the range (exclusive upper bound - 1ns)
		toInclusiveIST := params.ToDate.Add(-1 * time.Nanosecond).In(timeutil.IST)

		if fromIST.Year() == toInclusiveIST.Year() && fromIST.Month() == toInclusiveIST.Month() {
			m := fromIST.Format("2006-01")
			monthParam = &m
		} else {
			m := time.Now().In(timeutil.IST).Format("2006-01")
			monthParam = &m
		}
	} else if params.FromDate != nil || params.ToDate != nil {
		// If only one is provided, or if they are different months (handled above),
		// default to the current month as requested.
		m := time.Now().In(timeutil.IST).Format("2006-01")
		monthParam = &m
	}

	result, err := h.service.List(ctx, params, monthParam)
	if err != nil {
		if errors.Is(err, application.ErrInvalidMonthFormat) {
			writeErrorJSON(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "ERR_INVALID_MONTH")
			return
		}

		h.logger.ErrorContext(
			ctx,
			"failed to list allocations",
			slog.String("error", err.Error()),
		)
		writeErrorJSON(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list allocations", "ERR_INTERNAL_SERVER")
		return
	}

	items := make([]AllocationListItemResponse, len(result.Allocations))
	for i, item := range result.Allocations {
		items[i] = allocationListItemToResponse(item)
	}

	resp := ListAllocationsResponse{
		Allocations: items,
		Total:       result.Total,
		Limit:       limit,
		Offset:      offset,
		MonthlySummary: AllocationMonthlySummaryResponse{
			Month:             result.MonthlySummary.Month,
			TotalAllocated:    result.MonthlySummary.TotalAllocated,
			JarTotalAllocated: result.MonthlySummary.JarTotalAllocated,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
