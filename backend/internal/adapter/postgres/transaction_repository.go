// Package postgres provides PostgreSQL-backed implementations of the application's persistence ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// transactionRepository implements the TransactionRepository port using PostgreSQL.
type transactionRepository struct {
	db *pgxpool.Pool
}

// NewTransactionRepository creates and returns the transaction repository.
func NewTransactionRepository(db *pgxpool.Pool) ports.TransactionRepository {
	return &transactionRepository{db: db}
}

var _ ports.TransactionRepository = (*transactionRepository)(nil)

func transactionValues(t *transaction.Transaction) []any {
	return []any{
		t.ID,
		t.Name,
		t.Type,
		t.Category,
		t.FromAccountID,
		t.ToAccountID,
		t.JarID,
		t.Amount,
		t.OccurredAt,
		t.IsMasterIncome,
		t.CreatedAt,
		t.UpdatedAt,
	}
}

type transactionScan struct {
	fromAccountID *uuid.UUID
	toAccountID   *uuid.UUID
	jarID         *uuid.UUID
}

func transactionScanArgs(
	t *transaction.Transaction,
	scan *transactionScan,
) []any {
	return []any{
		&t.ID,
		&t.Name,
		&t.Type,
		&t.Category,
		&scan.fromAccountID,
		&scan.toAccountID,
		&scan.jarID,
		&t.Amount,
		&t.OccurredAt,
		&t.IsMasterIncome,
		&t.CreatedAt,
		&t.UpdatedAt,
	}
}

func applyTransactionScan(
	t *transaction.Transaction,
	scan *transactionScan,
) {
	t.FromAccountID = scan.fromAccountID
	t.ToAccountID = scan.toAccountID
	t.JarID = scan.jarID
}

func transactionListItemScanArgs(item *ports.TransactionListItem) []any {
	return []any{
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Amount,
		&item.JarName,
		&item.Account,
		&item.AccountBalance,
		&item.Category,
		&item.OccurredAt,
	}
}

const (
	insertTransactionQuery = `
		INSERT INTO transactions (
			id,
			name,
			type,
			category,
			from_account_id,
			to_account_id,
			jar_id,
			amount,
			occurred_at,
			is_master_income,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	findTransactionByIDQuery = `
		SELECT
			id,
			name,
			type,
			category,
			from_account_id,
			to_account_id,
			jar_id,
			amount,
			occurred_at,
			is_master_income,
			created_at,
			updated_at
		FROM transactions
		WHERE id = $1
	`

	saveTransactionQuery = `
		UPDATE transactions
		SET
			name = $2,
			type = $3,
			category = $4,
			from_account_id = $5,
			to_account_id = $6,
			jar_id = $7,
			amount = $8,
			occurred_at = $9,
			is_master_income = $10,
			updated_at = $11
		WHERE id = $1
	`

	deleteTransactionQuery = `DELETE FROM transactions WHERE id = $1`
)

func (r *transactionRepository) Create(
	ctx context.Context,
	t *transaction.Transaction,
) error {
	exec := dbExecutor(ctx, r.db)

	_, err := exec.Exec(
		ctx,
		insertTransactionQuery,
		transactionValues(t)...,
	)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	return nil
}

// List retrieves transactions with optional filtering and pagination.
// Filters are applied in PostgreSQL before pagination.
// No running balance is calculated (it's meaningless across multiple accounts).
func (r *transactionRepository) List(
	ctx context.Context,
	params ports.ListParams,
) ([]*ports.TransactionListItem, error) {
	exec := dbExecutor(ctx, r.db)

	query, queryParams := r.buildListQuery(params)

	rows, err := exec.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*ports.TransactionListItem, 0)

	for rows.Next() {
		var item ports.TransactionListItem

		if err := rows.Scan(
			transactionListItemScanArgs(&item)...,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}

		transactions = append(transactions, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	return transactions, nil
}

// buildListQuery constructs the SQL query and parameters for List() based on filters.
// This is the global transaction list without running balance.
func (r *transactionRepository) buildListQuery(
	params ports.ListParams,
) (string, []any) {
	var query strings.Builder
	var queryParams []any

	query.WriteString(`
		SELECT
			t.id,
			t.name,
			t.type,
			t.amount,
			j.name AS jar_name,
			CASE
				WHEN t.type = 'income' THEN to_account.name
				WHEN t.type = 'expense' THEN from_account.name
				WHEN t.type = 'transfer' THEN from_account.name || ' -> ' || to_account.name
				ELSE ''
			END AS account,
			CASE
				WHEN t.type = 'income' THEN to_account.balance
				WHEN t.type = 'expense' THEN from_account.balance
				ELSE NULL
			END AS account_balance,
			t.category,
			t.occurred_at
		FROM transactions t
		LEFT JOIN accounts from_account
			ON from_account.id = t.from_account_id
		LEFT JOIN accounts to_account
			ON to_account.id = t.to_account_id
		LEFT JOIN jars j
			ON j.id = t.jar_id
		WHERE 1 = 1
	`)

	if params.Search != nil && *params.Search != "" {
		query.WriteString(
			" AND t.name ILIKE '%' || $" +
				fmt.Sprintf("%d", len(queryParams)+1) +
				" || '%'",
		)
		queryParams = append(queryParams, *params.Search)
	}

	if params.AccountID != nil {
		n := len(queryParams) + 1

		query.WriteString(
			" AND (t.from_account_id = $" +
				fmt.Sprintf("%d", n) +
				" OR t.to_account_id = $" +
				fmt.Sprintf("%d", n) +
				")",
		)

		// Same PostgreSQL parameter is used for both comparisons.
		queryParams = append(queryParams, *params.AccountID)
	}

	if params.Type != nil {
		query.WriteString(
			" AND t.type = $" + fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.Type)
	}

	if params.Category != nil {
		query.WriteString(
			" AND t.category = $" + fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.Category)
	}

	if params.FromDate != nil {
		query.WriteString(
			" AND t.occurred_at >= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.FromDate)
	}

	if params.ToDate != nil {
		query.WriteString(
			" AND t.occurred_at <= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.ToDate)
	}

	if params.MinAmount != nil {
		query.WriteString(
			" AND t.amount >= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.MinAmount)
	}

	if params.MaxAmount != nil {
		query.WriteString(
			" AND t.amount <= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.MaxAmount)
	}

	query.WriteString(
		`
		ORDER BY t.occurred_at DESC, t.created_at DESC, t.id DESC
		LIMIT $` +
			fmt.Sprintf("%d", len(queryParams)+1) +
			` OFFSET $` +
			fmt.Sprintf("%d", len(queryParams)+2),
	)

	queryParams = append(
		queryParams,
		params.Limit,
		params.Offset,
	)

	return query.String(), queryParams
}

// ListByAccount retrieves transactions for a specific account with running balance.
func (r *transactionRepository) ListByAccount(
	ctx context.Context,
	accountID uuid.UUID,
	params ports.ListParams,
) ([]*ports.TransactionListItem, error) {
	exec := dbExecutor(ctx, r.db)

	query, queryParams := r.buildListByAccountQuery(
		accountID,
		params,
	)

	rows, err := exec.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, fmt.Errorf(
			"list transactions by account: %w",
			err,
		)
	}
	defer rows.Close()

	transactions := make([]*ports.TransactionListItem, 0)

	for rows.Next() {
		var item ports.TransactionListItem

		if err := rows.Scan(transactionListItemScanArgs(&item)...); err != nil {
			return nil, fmt.Errorf(
				"scan transaction with balance: %w",
				err,
			)
		}

		transactions = append(transactions, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"list transactions by account: %w",
			err,
		)
	}

	return transactions, nil
}

// buildListByAccountQuery constructs the SQL query with running balance.
func (r *transactionRepository) buildListByAccountQuery(
	accountID uuid.UUID,
	params ports.ListParams,
) (string, []any) {
	var query strings.Builder
	var queryParams []any

	// $1 is the account ID.
	queryParams = append(queryParams, accountID)

	query.WriteString(`
		WITH account_transactions AS (
			SELECT
				id,
				name,
				type,
				category,
				from_account_id,
				to_account_id,
				jar_id,
				amount,
				occurred_at,
				is_master_income,
				created_at,
				updated_at
			FROM transactions
			WHERE (
				from_account_id = $1
				OR to_account_id = $1
			)
		),
		with_running_balance AS (
			SELECT
				id,
				name,
				type,
				category,
				from_account_id,
				to_account_id,
				jar_id,
				amount,
				occurred_at,
				is_master_income,
				created_at,
				updated_at,
				COALESCE(
					SUM(
						CASE
							WHEN to_account_id = $1 THEN amount
							WHEN from_account_id = $1 THEN -amount
							ELSE 0
						END
					) OVER (
						ORDER BY
							occurred_at ASC,
							created_at ASC,
							id ASC
						ROWS BETWEEN
							UNBOUNDED PRECEDING AND CURRENT ROW
					),
					0
				) AS balance_after
			FROM account_transactions
		),
		filtered_transactions AS (
			SELECT *
			FROM with_running_balance
			WHERE 1 = 1
	`)

	if params.Search != nil && *params.Search != "" {
		query.WriteString(
			" AND name ILIKE '%' || $" +
				fmt.Sprintf("%d", len(queryParams)+1) +
				" || '%'",
		)
		queryParams = append(queryParams, *params.Search)
	}

	if params.Type != nil {
		query.WriteString(
			" AND type = $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.Type)
	}

	if params.Category != nil {
		query.WriteString(
			" AND category = $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.Category)
	}

	if params.FromDate != nil {
		query.WriteString(
			" AND occurred_at >= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.FromDate)
	}

	if params.ToDate != nil {
		query.WriteString(
			" AND occurred_at <= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.ToDate)
	}

	if params.MinAmount != nil {
		query.WriteString(
			" AND amount >= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.MinAmount)
	}

	if params.MaxAmount != nil {
		query.WriteString(
			" AND amount <= $" +
				fmt.Sprintf("%d", len(queryParams)+1),
		)
		queryParams = append(queryParams, *params.MaxAmount)
	}

	query.WriteString(`
		)
		SELECT
			id,
			name,
			type,
			CASE
				WHEN to_account_id = $1 THEN amount
				WHEN from_account_id = $1 THEN -amount
				ELSE amount
			END AS amount,
			jar_name,
			account,
			balance_after AS account_balance,
			category,
			occurred_at
		FROM (
			SELECT
				ft.*,
				j.name AS jar_name,
				CASE
					WHEN ft.type = 'transfer' AND ft.from_account_id = $1 THEN to_account.name
					WHEN ft.type = 'transfer' AND ft.to_account_id = $1 THEN from_account.name
					WHEN ft.type = 'income' THEN to_account.name
					WHEN ft.type = 'expense' THEN from_account.name
					ELSE ''
				END AS account
			FROM filtered_transactions ft
			LEFT JOIN accounts from_account
				ON from_account.id = ft.from_account_id
			LEFT JOIN accounts to_account
				ON to_account.id = ft.to_account_id
			LEFT JOIN jars j
				ON j.id = ft.jar_id
		) projected_transactions
		ORDER BY occurred_at DESC, created_at DESC, id DESC
		LIMIT $` +
		fmt.Sprintf("%d", len(queryParams)+1) +
		` OFFSET $` +
		fmt.Sprintf("%d", len(queryParams)+2),
	)

	queryParams = append(
		queryParams,
		params.Limit,
		params.Offset,
	)

	return query.String(), queryParams
}

func (r *transactionRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*transaction.Transaction, error) {
	exec := dbExecutor(ctx, r.db)

	var t transaction.Transaction
	var scan transactionScan

	err := exec.QueryRow(
		ctx,
		findTransactionByIDQuery,
		id,
	).Scan(transactionScanArgs(&t, &scan)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"find transaction: transaction not found",
			)
		}

		return nil, fmt.Errorf("find transaction: %w", err)
	}

	applyTransactionScan(&t, &scan)

	return &t, nil
}

func (r *transactionRepository) Save(
	ctx context.Context,
	t *transaction.Transaction,
) error {
	exec := dbExecutor(ctx, r.db)

	tag, err := exec.Exec(
		ctx,
		saveTransactionQuery,
		t.ID,
		t.Name,
		t.Type,
		t.Category,
		t.FromAccountID,
		t.ToAccountID,
		t.JarID,
		t.Amount,
		t.OccurredAt,
		t.IsMasterIncome,
		t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save transaction: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"save transaction: transaction not found",
		)
	}

	return nil
}

func (r *transactionRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	exec := dbExecutor(ctx, r.db)

	tag, err := exec.Exec(
		ctx,
		deleteTransactionQuery,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"delete transaction: transaction not found",
		)
	}

	return nil
}
