// Package postgres provides PostgreSQL-backed implementations of the application's persistence ports.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
		var createdAt time.Time // Ignore created_at in allocation struct

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
