package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/timeutil"
)

const (
	statementDefaultLimit = 20
	statementMaxLimit     = 100
)

// StatementService defines the interface required by the statement handler.
type StatementService interface {
	Get(
		ctx context.Context,
		accountID uuid.UUID,
		params ports.ListParams,
	) (*application.Statement, error)
}

// StatementHandler handles HTTP requests for account statements.
type StatementHandler struct {
	service StatementService
	logger  *slog.Logger
}

// NewStatementHandler creates a new StatementHandler.
func NewStatementHandler(
	service StatementService,
	logger *slog.Logger,
) *StatementHandler {
	return &StatementHandler{service: service, logger: logger}
}

// StatementAccountResponse is the account header of a statement.
type StatementAccountResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	IconKey    string    `json:"icon_key"`
	Balance    int64     `json:"balance"`
	IsPrimary  bool      `json:"is_primary"`
	IsArchived bool      `json:"is_archived"`
}

// StatementTransactionResponse is one statement row.
type StatementTransactionResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OccurredAt   time.Time `json:"occurred_at"`
	Type         string    `json:"type"`
	Category     string    `json:"category"`
	JarName      *string   `json:"jar_name"`
	Amount       int64     `json:"amount"`
	BalanceAfter *int64    `json:"balance_after"`
}

// StatementResponse is the body of GET /accounts/{id}/statement.
type StatementResponse struct {
	Account      StatementAccountResponse       `json:"account"`
	Transactions []StatementTransactionResponse `json:"transactions"`
	Total        int                            `json:"total"`
}

// HandleGet handles GET /accounts/{id}/statement.
//
// Query parameters (all optional):
//
//	from_date, to_date  plain date (IST day) or RFC 3339; both inclusive of
//	                    the whole day when a plain date is used. For a single
//	                    day pass the same date for both.
//	type                income | expense | transfer
//	category            transaction category
//	jar_id              jar UUID
//	search              case-insensitive substring of the transaction name
//	limit               1..100, default 20
//	offset              >= 0, default 0
func (h *StatementHandler) HandleGet(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "bad_request", "invalid account id", "INVALID_ACCOUNT_ID")
		return
	}

	params, code, msg := parseStatementParams(r)
	if code != "" {
		writeErrorJSON(w, http.StatusBadRequest, "bad_request", msg, code)
		return
	}

	stmt, err := h.service.Get(ctx, accountID, params)
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			writeErrorJSON(w, http.StatusNotFound, "not_found", "account not found", "ACCOUNT_NOT_FOUND")
			return
		}

		h.logger.ErrorContext(
			ctx,
			"failed to get statement",
			slog.String("account_id", accountID.String()),
			slog.String("error", err.Error()),
		)
		writeErrorJSON(w, http.StatusInternalServerError, "internal_error", "internal server error", "INTERNAL_ERROR")

		return
	}

	rows := make([]StatementTransactionResponse, len(stmt.Transactions))
	for i, item := range stmt.Transactions {
		rows[i] = StatementTransactionResponse{
			ID:           item.ID.String(),
			Name:         item.Name,
			OccurredAt:   item.OccurredAt.In(timeutil.IST),
			Type:         string(item.Type),
			Category:     string(item.Category),
			JarName:      item.JarName,
			Amount:       item.Amount,
			BalanceAfter: item.AccountBalance,
		}
	}

	writeJSON(w, http.StatusOK, StatementResponse{
		Account: StatementAccountResponse{
			ID:         stmt.Account.ID,
			Name:       stmt.Account.Name,
			IconKey:    stmt.Account.IconKey,
			Balance:    stmt.Account.Balance,
			IsPrimary:  stmt.Account.IsPrimary,
			IsArchived: stmt.Account.IsArchived,
		},
		Transactions: rows,
		Total:        stmt.Total,
	})
}

// parseStatementParams validates query parameters. Unlike the legacy list
// endpoints it rejects invalid values instead of silently ignoring them, so a
// typo in a filter can't show the user unfiltered data. On failure it returns
// a non-empty error code and message.
func parseStatementParams(
	r *http.Request,
) (params ports.ListParams, code string, msg string) {
	q := r.URL.Query()

	params.Limit = statementDefaultLimit

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > statementMaxLimit {
			return params, "INVALID_LIMIT", "limit must be between 1 and 100"
		}

		params.Limit = n
	}

	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return params, "INVALID_OFFSET", "offset must be a non-negative integer"
		}

		params.Offset = n
	}

	if v := q.Get("search"); v != "" {
		params.Search = &v
	}

	if v := q.Get("type"); v != "" {
		t := transaction.TransactionType(v)
		if !t.IsValid() {
			return params, "INVALID_TYPE", "type must be income, expense or transfer"
		}

		params.Type = &t
	}

	if v := q.Get("category"); v != "" {
		c := transaction.TransactionCategory(v)
		if !c.IsValid() {
			return params, "INVALID_CATEGORY", "unknown category"
		}

		params.Category = &c
	}

	if v := q.Get("jar_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return params, "INVALID_JAR_ID", "jar_id must be a UUID"
		}

		params.JarID = &id
	}

	if v := q.Get("from_date"); v != "" {
		t, ok := parseFilterTime(v, false)
		if !ok {
			return params, "INVALID_FROM_DATE", "from_date must be YYYY-MM-DD or RFC 3339"
		}

		params.FromDate = &t
	}

	if v := q.Get("to_date"); v != "" {
		t, ok := parseFilterTime(v, true)
		if !ok {
			return params, "INVALID_TO_DATE", "to_date must be YYYY-MM-DD or RFC 3339"
		}

		params.ToDate = &t
	}

	if params.FromDate != nil && params.ToDate != nil &&
		!params.FromDate.Before(*params.ToDate) {
		return params, "INVALID_DATE_RANGE", "from_date must not be after to_date"
	}

	return params, "", ""
}
