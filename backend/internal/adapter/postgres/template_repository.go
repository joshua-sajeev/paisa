// Package postgres provides PostgreSQL-backed implementations of the application's persistence ports.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

type templateRepository struct {
	db *pgxpool.Pool
}

func NewTemplateRepository(db *pgxpool.Pool) ports.TemplateRepository {
	return &templateRepository{db: db}
}

var _ ports.TemplateRepository = (*templateRepository)(nil)

func templateValues(t *transaction.Template) []any {
	return []any{
		t.ID,
		t.Name,
		t.Type,
		t.FromAccountID,
		t.ToAccountID,
		t.JarID,
		t.Amount,
		t.IsArchived,
		t.CreatedAt,
		t.UpdatedAt,
	}
}

func templateScanArgs(t *transaction.Template) []any {
	return []any{
		&t.ID,
		&t.Name,
		&t.Type,
		&t.FromAccountID,
		&t.ToAccountID,
		&t.JarID,
		&t.Amount,
		&t.IsArchived,
		&t.CreatedAt,
		&t.UpdatedAt,
	}
}

const (
	insertTemplateQuery = `
        INSERT INTO templates (
            id, name, type, from_account_id, to_account_id, jar_id, amount, is_archived, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `

	findTemplateByIDQuery = `
        SELECT id, name, type, from_account_id, to_account_id, jar_id, amount, is_archived, created_at, updated_at
        FROM templates WHERE id = $1
    `

	updateTemplateQuery = `
        UPDATE templates SET
            name = $2,
            type = $3,
            from_account_id = $4,
            to_account_id = $5,
            jar_id = $6,
            amount = $7,
            is_archived = $8,
            updated_at = $9
        WHERE id = $1
    `

	deleteTemplateQuery = `DELETE FROM templates WHERE id = $1`

	listTemplatesQuery = `
        SELECT id, name, type, from_account_id, to_account_id, jar_id, amount, is_archived, created_at, updated_at
        FROM templates
        ORDER BY created_at DESC, id DESC
        LIMIT $1 OFFSET $2
    `
)

func (r *templateRepository) Create(ctx context.Context, tmpl *transaction.Template) error {
	exec := dbExecutor(ctx, r.db)
	_, err := exec.Exec(ctx, insertTemplateQuery, templateValues(tmpl)...)
	if err != nil {
		// No unique constraints on template currently, just wrap error
		return fmt.Errorf("create template: %w", err)
	}
	return nil
}

func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*transaction.Template, error) {
	exec := dbExecutor(ctx, r.db)
	var t transaction.Template
	err := exec.QueryRow(ctx, findTemplateByIDQuery, id).Scan(templateScanArgs(&t)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("find template: %w", transaction.ErrTemplateNotFound)
		}
		return nil, fmt.Errorf("find template: %w", err)
	}
	return &t, nil
}

func (r *templateRepository) Update(ctx context.Context, tmpl *transaction.Template) error {
	exec := dbExecutor(ctx, r.db)
	tag, err := exec.Exec(ctx, updateTemplateQuery, tmpl.ID, tmpl.Name, tmpl.Type, tmpl.FromAccountID, tmpl.ToAccountID, tmpl.JarID, tmpl.Amount, tmpl.IsArchived, time.Now().UTC().Truncate(time.Microsecond))
	if err != nil {
		return fmt.Errorf("update template: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update template: %w", transaction.ErrTemplateNotFound)
	}
	return nil
}

func (r *templateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	exec := dbExecutor(ctx, r.db)
	tag, err := exec.Exec(ctx, deleteTemplateQuery, id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete template: %w", transaction.ErrTemplateNotFound)
	}
	return nil
}

func (r *templateRepository) List(ctx context.Context, opts ports.ListOptions) ([]*transaction.Template, error) {
	exec := dbExecutor(ctx, r.db)
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := opts.Offset
	rows, err := exec.Query(ctx, listTemplatesQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()
	var results []*transaction.Template
	for rows.Next() {
		var t transaction.Template
		if err := rows.Scan(templateScanArgs(&t)...); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		results = append(results, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list templates rows: %w", err)
	}
	return results, nil
}
