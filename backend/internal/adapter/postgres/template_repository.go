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
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

type templateRepository struct{ db *pgxpool.Pool }

func NewTemplateRepository(db *pgxpool.Pool) ports.TemplateRepository {
	return &templateRepository{db: db}
}

var _ ports.TemplateRepository = (*templateRepository)(nil)

func templateValues(t *transaction.Template) []any {
	return []any{t.ID, t.Name, t.Type, t.Category, t.FromAccountID, t.ToAccountID, t.JarID, t.Amount, t.IsArchived, t.CreatedAt, t.UpdatedAt}
}

type templateNullableIDs struct{ from, to, jar *uuid.UUID }

func templateScanArgs(t *transaction.Template, n *templateNullableIDs) []any {
	return []any{&t.ID, &t.Name, &t.Type, &t.Category, &n.from, &n.to, &n.jar, &t.Amount, &t.IsArchived, &t.CreatedAt, &t.UpdatedAt}
}
func applyTemplateScan(t *transaction.Template, n templateNullableIDs) {
	t.FromAccountID, t.ToAccountID, t.JarID = n.from, n.to, n.jar
}

const templateColumns = `id, name, type, category, from_account_id, to_account_id, jar_id, amount, is_archived, created_at, updated_at`
const insertTemplateQuery = `INSERT INTO templates (` + templateColumns + `) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
const findTemplateByIDQuery = `SELECT ` + templateColumns + ` FROM templates WHERE id = $1`
const updateTemplateQuery = `UPDATE templates SET name=$2, type=$3, category=$4, from_account_id=$5, to_account_id=$6, jar_id=$7, amount=$8, is_archived=$9, updated_at=$10 WHERE id=$1`
const deleteTemplateQuery = `DELETE FROM templates WHERE id=$1`
const listTemplatesQuery = `SELECT ` + templateColumns + ` FROM templates WHERE is_archived=$1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`

func (r *templateRepository) Create(ctx context.Context, t *transaction.Template) error {
	_, err := dbExecutor(ctx, r.db).Exec(ctx, insertTemplateQuery, templateValues(t)...)
	if err != nil {
		return templateWriteError("create template", err)
	}
	return nil
}
func (r *templateRepository) GetByID(ctx context.Context, id uuid.UUID) (*transaction.Template, error) {
	var t transaction.Template
	var n templateNullableIDs
	err := dbExecutor(ctx, r.db).QueryRow(ctx, findTemplateByIDQuery, id).Scan(templateScanArgs(&t, &n)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("find template: %w", transaction.ErrTemplateNotFound)
		}
		return nil, fmt.Errorf("find template: %w", err)
	}
	applyTemplateScan(&t, n)
	return &t, nil
}
func (r *templateRepository) Update(ctx context.Context, t *transaction.Template) error {
	tag, err := dbExecutor(ctx, r.db).Exec(ctx, updateTemplateQuery, t.ID, t.Name, t.Type, t.Category, t.FromAccountID, t.ToAccountID, t.JarID, t.Amount, t.IsArchived, time.Now().UTC().Truncate(time.Microsecond))
	if err != nil {
		return templateWriteError("update template", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update template: %w", transaction.ErrTemplateNotFound)
	}
	return nil
}
func (r *templateRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := dbExecutor(ctx, r.db).Exec(ctx, deleteTemplateQuery, id)
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete template: %w", transaction.ErrTemplateNotFound)
	}
	return nil
}
func (r *templateRepository) List(ctx context.Context, opts ports.ListOptions) ([]*transaction.Template, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := dbExecutor(ctx, r.db).Query(ctx, listTemplatesQuery, opts.IsArchived, limit, max(0, opts.Offset))
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()
	results := make([]*transaction.Template, 0)
	for rows.Next() {
		var t transaction.Template
		var n templateNullableIDs
		if err := rows.Scan(templateScanArgs(&t, &n)...); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		applyTemplateScan(&t, n)
		results = append(results, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list templates rows: %w", err)
	}
	return results, nil
}
func templateWriteError(operation string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%s: %w", operation, transaction.ErrTemplateNameExists)
	}
	return fmt.Errorf("%s: %w", operation, err)
}
