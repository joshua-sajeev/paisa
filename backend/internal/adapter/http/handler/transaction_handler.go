package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// TransactionService defines the interface required by the handler.
type TransactionService interface {
	Create(
		ctx context.Context,
		name string,
		transactionType transaction.TransactionType,
		category transaction.TransactionCategory,
		fromAccountID *uuid.UUID,
		toAccountID *uuid.UUID,
		jarID *uuid.UUID,
		amount int64,
		occurredAt time.Time,
		isMasterIncome bool,
	) (*transaction.Transaction, error)

	Update(
		ctx context.Context,
		id uuid.UUID,
		name string,
		category transaction.TransactionCategory,
		fromAccountID *uuid.UUID,
		toAccountID *uuid.UUID,
		jarID *uuid.UUID,
		amount int64,
		occurredAt time.Time,
		isMasterIncome bool,
	) (*transaction.Transaction, error)

	Delete(ctx context.Context, id uuid.UUID) error

	// List retrieves transactions with optional filtering and pagination.
	List(
		ctx context.Context,
		params ports.ListParams,
	) ([]*ports.TransactionListItem, error)

	// ListByAccount retrieves transactions for a specific account with running balance.
	ListByAccount(
		ctx context.Context,
		accountID uuid.UUID,
		params ports.ListParams,
	) ([]*ports.TransactionListItem, error)

	GetByID(ctx context.Context, id uuid.UUID) (*transaction.Transaction, error)
}

// TransactionHandler handles HTTP requests for transactions.
type TransactionHandler struct {
	service TransactionService
	logger  *slog.Logger
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(
	service TransactionService,
	logger *slog.Logger,
) *TransactionHandler {
	return &TransactionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateTransactionRequest represents the request payload for creating a transaction.
type CreateTransactionRequest struct {
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Category       string    `json:"category"`
	FromAccountID  string    `json:"from_account_id"`
	ToAccountID    string    `json:"to_account_id"`
	JarID          string    `json:"jar_id"`
	Amount         int64     `json:"amount"`
	OccurredAt     time.Time `json:"occurred_at"`
	IsMasterIncome bool      `json:"is_master_income"`
}

// UpdateTransactionRequest represents the request payload for updating a transaction.
type UpdateTransactionRequest struct {
	Name           string    `json:"name"`
	Category       string    `json:"category"`
	FromAccountID  string    `json:"from_account_id"`
	ToAccountID    string    `json:"to_account_id"`
	JarID          string    `json:"jar_id"`
	Amount         int64     `json:"amount"`
	OccurredAt     time.Time `json:"occurred_at"`
	IsMasterIncome bool      `json:"is_master_income"`
}

// TransactionResponse represents a transaction in the response.
type TransactionResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Category       string    `json:"category"`
	FromAccountID  *string   `json:"from_account_id"`
	ToAccountID    *string   `json:"to_account_id"`
	JarID          *string   `json:"jar_id"`
	Amount         int64     `json:"amount"`
	OccurredAt     time.Time `json:"occurred_at"`
	IsMasterIncome bool      `json:"is_master_income"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	BalanceAfter   *int64    `json:"balance_after"`
}

func transactionToResponse(
	t *transaction.Transaction,
	balance *int64,
) TransactionResponse {
	resp := TransactionResponse{
		ID:             t.ID.String(),
		Name:           t.Name,
		Type:           string(t.Type),
		Category:       string(t.Category),
		Amount:         t.Amount,
		OccurredAt:     t.OccurredAt,
		IsMasterIncome: t.IsMasterIncome,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		BalanceAfter:   balance,
	}

	if t.FromAccountID != nil {
		id := t.FromAccountID.String()
		resp.FromAccountID = &id
	}

	if t.ToAccountID != nil {
		id := t.ToAccountID.String()
		resp.ToAccountID = &id
	}

	if t.JarID != nil {
		id := t.JarID.String()
		resp.JarID = &id
	}

	return resp
}

// TransactionListItemResponse represents a transaction in list responses.
type TransactionListItemResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Date           string  `json:"date"`
	Type           string  `json:"type"`
	Amount         int64   `json:"amount"`
	JarName        *string `json:"jar_name"`
	Account        string  `json:"account"`
	AccountBalance *int64  `json:"account_balance"`
	Category       string  `json:"category"`
}

func transactionListItemToResponse(
	item *ports.TransactionListItem,
) TransactionListItemResponse {
	return TransactionListItemResponse{
		ID:             item.ID.String(),
		Name:           item.Name,
		Date:           item.OccurredAt.Format(time.DateOnly),
		Type:           string(item.Type),
		Amount:         item.Amount,
		JarName:        item.JarName,
		Account:        item.Account,
		AccountBalance: item.AccountBalance,
		Category:       string(item.Category),
	}
}

// parseOptionalUUID parses an optional UUID string.
// An empty string is represented as nil so that it can be stored as SQL NULL.
func parseOptionalUUID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

// HandleCreate handles POST /transactions.
func (h *TransactionHandler) HandleCreate(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(
			ctx,
			"invalid request body",
			slog.String("error", err.Error()),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	fromAcctID, err := parseOptionalUUID(req.FromAccountID)
	if err != nil {
		http.Error(w, "invalid from_account_id", http.StatusBadRequest)
		return
	}

	toAcctID, err := parseOptionalUUID(req.ToAccountID)
	if err != nil {
		http.Error(w, "invalid to_account_id", http.StatusBadRequest)
		return
	}

	jarID, err := parseOptionalUUID(req.JarID)
	if err != nil {
		http.Error(w, "invalid jar_id", http.StatusBadRequest)
		return
	}

	txnType := transaction.TransactionType(req.Type)
	category := transaction.TransactionCategory(req.Category)

	txn, err := h.service.Create(
		ctx,
		req.Name,
		txnType,
		category,
		fromAcctID,
		toAcctID,
		jarID,
		req.Amount,
		req.OccurredAt,
		req.IsMasterIncome,
	)
	if err != nil {
		h.logger.WarnContext(
			ctx,
			"failed to create transaction",
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(transactionToResponse(txn, nil))
}

// ListTransactionsResponse represents the global transaction list response.
type ListTransactionsResponse struct {
	Transactions []TransactionListItemResponse `json:"transactions"`
	Total        int                           `json:"total"`
}

// HandleList handles GET /transactions.
// Supports filtering by search, account, type, category, date range, and amount range.
func (h *TransactionHandler) HandleList(
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

	params := ports.ListParams{
		Limit:  limit,
		Offset: offset,
	}

	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = &search
	}

	if accountIDStr := r.URL.Query().Get("account_id"); accountIDStr != "" {
		accountID, err := uuid.Parse(accountIDStr)
		if err == nil {
			params.AccountID = &accountID
		}
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		txnType := transaction.TransactionType(typeStr)
		if txnType.IsValid() {
			params.Type = &txnType
		}
	}

	if categoryStr := r.URL.Query().Get("category"); categoryStr != "" {
		category := transaction.TransactionCategory(categoryStr)
		if category.IsValid() {
			params.Category = &category
		}
	}

	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.DateOnly, fromDateStr)
		if err == nil {
			params.FromDate = &fromDate
		}
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.DateOnly, toDateStr)
		if err == nil {
			params.ToDate = &toDate
		}
	}

	if minAmountStr := r.URL.Query().Get("min_amount"); minAmountStr != "" {
		minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
		if err == nil {
			params.MinAmount = &minAmount
		}
	}

	if maxAmountStr := r.URL.Query().Get("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
		if err == nil {
			params.MaxAmount = &maxAmount
		}
	}

	txns, err := h.service.List(ctx, params)
	if err != nil {
		h.logger.ErrorContext(
			ctx,
			"failed to list transactions",
			slog.String("error", err.Error()),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	responses := make([]TransactionListItemResponse, len(txns))
	for i, txn := range txns {
		responses[i] = transactionListItemToResponse(txn)
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(ListTransactionsResponse{
		Transactions: responses,
		Total:        len(txns),
	})
}

// ListAccountTransactionsResponse represents the account statement response.
type ListAccountTransactionsResponse struct {
	Transactions []TransactionListItemResponse `json:"transactions"`
	Total        int                           `json:"total"`
}

// HandleListByAccount handles GET /accounts/{id}/transactions.
// Returns transactions for the specified account with running balance.
func (h *TransactionHandler) HandleListByAccount(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	accountIDStr := chi.URLParam(r, "id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		http.Error(w, "invalid account id", http.StatusBadRequest)
		return
	}

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

	params := ports.ListParams{
		Limit:  limit,
		Offset: offset,
	}

	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = &search
	}

	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		txnType := transaction.TransactionType(typeStr)
		if txnType.IsValid() {
			params.Type = &txnType
		}
	}

	if categoryStr := r.URL.Query().Get("category"); categoryStr != "" {
		category := transaction.TransactionCategory(categoryStr)
		if category.IsValid() {
			params.Category = &category
		}
	}

	if fromDateStr := r.URL.Query().Get("from_date"); fromDateStr != "" {
		fromDate, err := time.Parse(time.DateOnly, fromDateStr)
		if err == nil {
			params.FromDate = &fromDate
		}
	}

	if toDateStr := r.URL.Query().Get("to_date"); toDateStr != "" {
		toDate, err := time.Parse(time.DateOnly, toDateStr)
		if err == nil {
			params.ToDate = &toDate
		}
	}

	if minAmountStr := r.URL.Query().Get("min_amount"); minAmountStr != "" {
		minAmount, err := strconv.ParseInt(minAmountStr, 10, 64)
		if err == nil {
			params.MinAmount = &minAmount
		}
	}

	if maxAmountStr := r.URL.Query().Get("max_amount"); maxAmountStr != "" {
		maxAmount, err := strconv.ParseInt(maxAmountStr, 10, 64)
		if err == nil {
			params.MaxAmount = &maxAmount
		}
	}

	txns, err := h.service.ListByAccount(
		ctx,
		accountID,
		params,
	)
	if err != nil {
		h.logger.ErrorContext(
			ctx,
			"failed to list transactions by account",
			slog.String("account_id", accountID.String()),
			slog.String("error", err.Error()),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	responses := make([]TransactionListItemResponse, len(txns))
	for i, txn := range txns {
		responses[i] = transactionListItemToResponse(txn)
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(ListAccountTransactionsResponse{
		Transactions: responses,
		Total:        len(txns),
	})
}

// HandleGetByID handles GET /transactions/{id}.
func (h *TransactionHandler) HandleGetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	txn, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.logger.WarnContext(
			ctx,
			"transaction not found",
			slog.String("error", err.Error()),
		)
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(transactionToResponse(txn, nil))
}

// HandleUpdate handles PATCH /transactions/{id}.
func (h *TransactionHandler) HandleUpdate(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WarnContext(
			ctx,
			"invalid request body",
			slog.String("error", err.Error()),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	fromAcctID, err := parseOptionalUUID(req.FromAccountID)
	if err != nil {
		http.Error(w, "invalid from_account_id", http.StatusBadRequest)
		return
	}

	toAcctID, err := parseOptionalUUID(req.ToAccountID)
	if err != nil {
		http.Error(w, "invalid to_account_id", http.StatusBadRequest)
		return
	}

	jarID, err := parseOptionalUUID(req.JarID)
	if err != nil {
		http.Error(w, "invalid jar_id", http.StatusBadRequest)
		return
	}

	category := transaction.TransactionCategory(req.Category)

	txn, err := h.service.Update(
		ctx,
		id,
		req.Name,
		category,
		fromAcctID,
		toAcctID,
		jarID,
		req.Amount,
		req.OccurredAt,
		req.IsMasterIncome,
	)
	if err != nil {
		h.logger.WarnContext(
			ctx,
			"failed to update transaction",
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(transactionToResponse(txn, nil))
}

// HandleDelete handles DELETE /transactions/{id}.
func (h *TransactionHandler) HandleDelete(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.WarnContext(
			ctx,
			"failed to delete transaction",
			slog.String("error", err.Error()),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
