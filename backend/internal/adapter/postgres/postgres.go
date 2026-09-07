package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

type txKey struct{}

// dbExec interface matches methods shared by *pgxpool.Pool and pgx.Tx.
type dbExec interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// dbExecutor returns either the active transaction from the context
// or the connection pool.
func dbExecutor(ctx context.Context, pool *pgxpool.Pool) dbExec {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return pool
}

// TxManager manages PostgreSQL transactions.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a new PostgreSQL transaction manager.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{
		pool: pool,
	}
}

// WithinTransaction executes fn inside a PostgreSQL transaction.
//
// The transaction is committed when fn returns nil. If fn returns an
// error, the transaction is rolled back and the original error is returned.
func (m *TxManager) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil &&
			!errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return fmt.Errorf(
				"transaction failed: %w; rollback failed: %v",
				err,
				rollbackErr,
			)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

var _ ports.TxManager = (*TxManager)(nil)
