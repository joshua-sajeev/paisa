// Package postgres provides PostgreSQL-backed implementations of the application's persistence ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// accountRepository implements the AccountRepository port using PostgreSQL.
type accountRepository struct {
	db *pgxpool.Pool
}

// NewAccountRepository creates and returns the interface
func NewAccountRepository(db *pgxpool.Pool) ports.AccountRepository {
	return &accountRepository{db: db}
}

var _ ports.AccountRepository = (*accountRepository)(nil)

type txKey struct{}

// dbExec interface matches methods shared by *pgxpool.Pool and pgx.Tx
type dbExec interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// dbExecutor returns either the active transaction from the context or the connection pool.
func dbExecutor(ctx context.Context, pool *pgxpool.Pool) dbExec {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

func accountValues(a *account.Account) []any {
	return []any{
		a.ID,
		a.Name,
		a.IsArchived,
		a.CreatedAt,
		a.UpdatedAt,
	}
}

func accountScanArgs(a *account.Account) []any {
	return []any{
		&a.ID,
		&a.Name,
		&a.IsArchived,
		&a.CreatedAt,
		&a.UpdatedAt,
	}
}

const (
	insertAccountQuery = `
		INSERT INTO accounts (
			id,
			name,
			is_archived,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	listAccountsQuery = `
		SELECT
			id,
			name,
			is_archived,
			created_at,
			updated_at
		FROM accounts
		ORDER BY created_at DESC, id DESC
	`

	updateAccountNameQuery = `
		UPDATE accounts
		SET
			name = $2,
			updated_at = $3
		WHERE id = $1
	`
	setAccountArchivedQuery = `
		UPDATE accounts
		SET
			is_archived = $2,
			updated_at = $3
		WHERE id = $1
	`
)

func (r *accountRepository) Create(ctx context.Context, a *account.Account) error {
	exec := dbExecutor(ctx, r.db)

	_, err := exec.Exec(ctx, insertAccountQuery, accountValues(a)...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("create account: %w", account.ErrAccountNameExists)
		}

		return fmt.Errorf("create account: %w", err)
	}

	return nil
}

func (r *accountRepository) List(ctx context.Context) ([]*account.Account, error) {
	exec := dbExecutor(ctx, r.db)

	rows, err := exec.Query(ctx, listAccountsQuery)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]*account.Account, 0)

	for rows.Next() {
		var a account.Account

		if err := rows.Scan(accountScanArgs(&a)...); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}

		accounts = append(accounts, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	return accounts, nil
}

func (r *accountRepository) UpdateName(ctx context.Context, id uuid.UUID, name string) error {
	exec := dbExecutor(ctx, r.db)

	cmd, err := exec.Exec(
		ctx,
		updateAccountNameQuery,
		id,
		name,
		time.Now().UTC().Truncate(time.Microsecond),
	)
	if err != nil {
		return fmt.Errorf("update account name: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return account.ErrAccountNotFound
	}

	return nil
}

func (r *accountRepository) SetArchived(ctx context.Context, id uuid.UUID, archived bool) error {
	exec := dbExecutor(ctx, r.db)

	now := time.Now().UTC().Truncate(time.Microsecond)

	tag, err := exec.Exec(ctx, setAccountArchivedQuery, id, archived, now)
	if err != nil {
		return fmt.Errorf("set account archived: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("set account archived: %w", account.ErrAccountNotFound)
	}

	return nil
}
