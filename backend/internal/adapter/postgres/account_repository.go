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

// NewAccountRepository creates and returns the account repository.
func NewAccountRepository(db *pgxpool.Pool) ports.AccountRepository {
	return &accountRepository{db: db}
}

var _ ports.AccountRepository = (*accountRepository)(nil)

func accountValues(a *account.Account) []any {
	return []any{
		a.ID,
		a.Name,
		a.IconKey,
		a.Balance,
		a.IsArchived,
		a.CreatedAt,
		a.UpdatedAt,
	}
}

func accountScanArgs(a *account.Account) []any {
	return []any{
		&a.ID,
		&a.Name,
		&a.IconKey,
		&a.Balance,
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
			icon_key,
			balance,
			is_archived,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	listAccountsQuery = `
		SELECT
			id,
			name,
			icon_key,
			balance,
			is_archived,
			created_at,
			updated_at
		FROM accounts
		ORDER BY created_at DESC, id DESC
	`

	findAccountByIDQuery = `
		SELECT
			id,
			name,
			icon_key,
			balance,
			is_archived,
			created_at,
			updated_at
		FROM accounts
		WHERE id = $1
	`

	saveAccountQuery = `
		UPDATE accounts
		SET
			name = $2,
			icon_key = $3,
			is_archived = $4,
			updated_at = $5
		WHERE id = $1
	`

	adjustAccountBalanceQuery = `
		UPDATE accounts
		SET
			balance = balance + $2,
			updated_at = $3
		WHERE id = $1
	`
)

func (r *accountRepository) Create(ctx context.Context, a *account.Account) error {
	exec := dbExecutor(ctx, r.db)

	_, err := exec.Exec(
		ctx,
		insertAccountQuery,
		accountValues(a)...,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf(
				"create account: %w",
				account.ErrAccountNameExists,
			)
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

func (r *accountRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*account.Account, error) {
	exec := dbExecutor(ctx, r.db)

	var a account.Account

	err := exec.QueryRow(
		ctx,
		findAccountByIDQuery,
		id,
	).Scan(accountScanArgs(&a)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"find account: %w",
				account.ErrAccountNotFound,
			)
		}

		return nil, fmt.Errorf("find account: %w", err)
	}

	return &a, nil
}

func (r *accountRepository) Save(
	ctx context.Context,
	a *account.Account,
) error {
	exec := dbExecutor(ctx, r.db)

	tag, err := exec.Exec(
		ctx,
		saveAccountQuery,
		a.ID,
		a.Name,
		a.IconKey,
		a.IsArchived,
		a.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf(
				"save account: %w",
				account.ErrAccountNameExists,
			)
		}

		return fmt.Errorf("save account: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"save account: %w",
			account.ErrAccountNotFound,
		)
	}

	return nil
}

func (r *accountRepository) AdjustBalance(
	ctx context.Context,
	id uuid.UUID,
	delta int64,
) error {
	exec := dbExecutor(ctx, r.db)

	tag, err := exec.Exec(
		ctx,
		adjustAccountBalanceQuery,
		id,
		delta,
		time.Now().UTC().Truncate(time.Microsecond),
	)
	if err != nil {
		return fmt.Errorf("adjust account balance: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"adjust account balance: %w",
			account.ErrAccountNotFound,
		)
	}

	return nil
}
