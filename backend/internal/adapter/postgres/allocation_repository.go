// Package postgres provides PostgreSQL-backed implementations of the application's persistence ports.
package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// allocationRepository implements the AllocationRepository port using PostgreSQL.
type allocationRepository struct {
	db *pgxpool.Pool
}

// NewAllocationRepository creates and returns the allocation repository.
func NewAllocationRepository(db *pgxpool.Pool) ports.AllocationRepository {
	return &allocationRepository{db: db}
}

var _ ports.AllocationRepository = (*allocationRepository)(nil)

const (
	insertAllocationQuery = `
		INSERT INTO jar_allocations (
			id,
			transaction_id,
			jar_id,
			amount,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	deleteByTransactionIDQuery = `
		DELETE FROM jar_allocations
		WHERE transaction_id = $1
	`

	listByTransactionIDQuery = `
		SELECT
			id,
			transaction_id,
			jar_id,
			amount,
			created_at
		FROM jar_allocations
		WHERE transaction_id = $1
		ORDER BY created_at ASC
	`
	sumByJarQuery = `
		SELECT COALESCE(SUM(a.amount), 0)
		FROM jar_allocations a
		INNER JOIN transactions t ON t.id = a.transaction_id
		WHERE a.jar_id = $1
	`
)

func (r *allocationRepository) Create(
	ctx context.Context,
	alloc *transaction.Allocation,
) error {
	exec := dbExecutor(ctx, r.db)

	now := time.Now().UTC().Truncate(time.Microsecond)

	_, err := exec.Exec(
		ctx,
		insertAllocationQuery,
		alloc.ID,
		alloc.TransactionID,
		alloc.JarID,
		alloc.Amount,
		now,
	)
	if err != nil {
		return fmt.Errorf("create allocation: %w", err)
	}

	return nil
}

func (r *allocationRepository) DeleteByTransactionID(
	ctx context.Context,
	transactionID uuid.UUID,
) error {
	exec := dbExecutor(ctx, r.db)

	_, err := exec.Exec(ctx, deleteByTransactionIDQuery, transactionID)
	if err != nil {
		return fmt.Errorf("delete allocations by transaction: %w", err)
	}

	return nil
}

func (r *allocationRepository) ListByTransactionID(
	ctx context.Context,
	transactionID uuid.UUID,
) ([]*transaction.Allocation, error) {
	exec := dbExecutor(ctx, r.db)

	rows, err := exec.Query(ctx, listByTransactionIDQuery, transactionID)
	if err != nil {
		return nil, fmt.Errorf("list allocations by transaction: %w", err)
	}
	defer rows.Close()

	allocations := make([]*transaction.Allocation, 0)

	for rows.Next() {
		var alloc transaction.Allocation
		var createdAt time.Time

		if err := rows.Scan(
			&alloc.ID,
			&alloc.TransactionID,
			&alloc.JarID,
			&alloc.Amount,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan allocation: %w", err)
		}

		allocations = append(allocations, &alloc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list allocations by transaction: %w", err)
	}

	return allocations, nil
}

func (r *allocationRepository) buildFilterConditions(
	params ports.AllocationListParams,
) (string, []any) {
	var sb strings.Builder
	var args []any

	if params.Search != nil && *params.Search != "" {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND (t.name ILIKE '%%' || $%d || '%%' OR j.name ILIKE '%%' || $%d || '%%')", n, n))
		args = append(args, *params.Search)
	}

	if params.JarID != nil {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND a.jar_id = $%d", n))
		args = append(args, *params.JarID)
	}

	if params.JarName != nil && *params.JarName != "" {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND j.name ILIKE '%%' || $%d || '%%'", n))
		args = append(args, *params.JarName)
	}

	if params.TransactionName != nil && *params.TransactionName != "" {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND t.name ILIKE '%%' || $%d || '%%'", n))
		args = append(args, *params.TransactionName)
	}

	if params.AllocationType != nil && params.AllocationType.IsValid() {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND j.allocation_type = $%d", n))
		args = append(args, string(*params.AllocationType))
	}

	if params.FromDate != nil {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND t.occurred_at >= $%d", n))
		args = append(args, *params.FromDate)
	}

	if params.ToDate != nil {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND t.occurred_at < $%d", n))
		args = append(args, *params.ToDate)
	}

	if params.MinAmount != nil {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND a.amount >= $%d", n))
		args = append(args, *params.MinAmount)
	}

	if params.MaxAmount != nil {
		n := len(args) + 1
		sb.WriteString(fmt.Sprintf(" AND a.amount <= $%d", n))
		args = append(args, *params.MaxAmount)
	}

	return sb.String(), args
}

func (r *allocationRepository) List(
	ctx context.Context,
	params ports.AllocationListParams,
) ([]*ports.AllocationListItem, int, error) {
	exec := dbExecutor(ctx, r.db)

	whereClause, args := r.buildFilterConditions(params)

	countQuery := `
		SELECT COUNT(*)
		FROM jar_allocations a
		INNER JOIN transactions t ON t.id = a.transaction_id
		INNER JOIN jars j ON j.id = a.jar_id
		WHERE 1 = 1
	` + whereClause

	var total int
	if err := exec.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count allocations: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT
			a.id,
			a.transaction_id,
			t.name AS transaction_name,
			a.jar_id,
			j.name AS jar_name,
			j.allocation_type,
			a.amount,
			t.occurred_at,
			a.created_at
		FROM jar_allocations a
		INNER JOIN transactions t ON t.id = a.transaction_id
		INNER JOIN jars j ON j.id = a.jar_id
		WHERE 1 = 1 %s
		ORDER BY t.occurred_at DESC, a.created_at DESC, a.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	queryArgs := append(args, limit, offset)

	rows, err := exec.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list allocations: %w", err)
	}
	defer rows.Close()

	items := make([]*ports.AllocationListItem, 0)

	for rows.Next() {
		var item ports.AllocationListItem
		var allocType string

		if err := rows.Scan(
			&item.ID,
			&item.TransactionID,
			&item.TransactionName,
			&item.JarID,
			&item.JarName,
			&allocType,
			&item.Amount,
			&item.OccurredAt,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan allocation item: %w", err)
		}

		item.AllocationType = jar.AllocationType(allocType)
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list allocations: %w", err)
	}

	return items, total, nil
}

func (r *allocationRepository) GetMonthlySummary(
	ctx context.Context,
	start time.Time,
	end time.Time,
	jarID *uuid.UUID,
) (int64, error) {
	exec := dbExecutor(ctx, r.db)

	query := `
		SELECT COALESCE(SUM(a.amount), 0)
		FROM jar_allocations a
		INNER JOIN transactions t ON t.id = a.transaction_id
		WHERE t.occurred_at >= $1 AND t.occurred_at < $2
	`
	args := []any{start, end}

	if jarID != nil {
		query += " AND a.jar_id = $3"
		args = append(args, *jarID)
	}

	var total int64
	if err := exec.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("get monthly allocation summary: %w", err)
	}

	return total, nil
}

func (r *allocationRepository) SumByJar(
	ctx context.Context,
	jarID uuid.UUID,
	startDate *time.Time,
	endDate *time.Time,
) (int64, error) {
	exec := dbExecutor(ctx, r.db)

	query := sumByJarQuery
	args := []any{jarID}
	argIndex := 2

	if startDate != nil {
		query += fmt.Sprintf(
			" AND t.occurred_at >= $%d",
			argIndex,
		)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		query += fmt.Sprintf(
			" AND t.occurred_at < $%d",
			argIndex,
		)
		args = append(args, *endDate)
	}

	var total int64

	if err := exec.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("sum allocations by jar: %w", err)
	}

	return total, nil
}
