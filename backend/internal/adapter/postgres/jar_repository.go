// Package postgres provides PostgreSQL-backed implementations of the application's persistence ports.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// jarRepository implements the JarRepository port using PostgreSQL.
type jarRepository struct {
	db *pgxpool.Pool
}

// NewJarRepository creates and returns the jar repository.
func NewJarRepository(db *pgxpool.Pool) ports.JarRepository {
	return &jarRepository{db: db}
}

var _ ports.JarRepository = (*jarRepository)(nil)

func jarValues(j *jar.Jar) []any {
	return []any{
		j.ID,
		j.Name,
		j.AllocationType,
		j.AllocationValue,
		j.IsArchived,
		j.CreatedAt,
		j.UpdatedAt,
	}
}

func jarSaveValues(j *jar.Jar) []any {
	return []any{
		j.ID,
		j.Name,
		j.AllocationType,
		j.AllocationValue,
		j.IsArchived,
		j.UpdatedAt,
	}
}

func jarScanArgs(j *jar.Jar) []any {
	return []any{
		&j.ID,
		&j.Name,
		&j.AllocationType,
		&j.AllocationValue,
		&j.IsArchived,
		&j.CreatedAt,
		&j.UpdatedAt,
	}
}

const (
	insertJarQuery = `
		INSERT INTO jars (
			id,
			name,
			allocation_type,
			allocation_value,
			is_archived,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	listJarsQuery = `
		SELECT
			id,
			name,
			allocation_type,
			allocation_value,
			is_archived,
			created_at,
			updated_at
		FROM jars
		ORDER BY created_at DESC, id DESC
	`

	findJarByIDQuery = `
		SELECT
			id,
			name,
			allocation_type,
			allocation_value,
			is_archived,
			created_at,
			updated_at
		FROM jars
		WHERE id = $1
	`

	saveJarQuery = `
		UPDATE jars
		SET
			name = $2,
			allocation_type = $3,
			allocation_value = $4,
			is_archived = $5,
			updated_at = $6
		WHERE id = $1
	`

	updateJarAllocationQuery = `
		UPDATE jars
		SET
			allocation_type = $2,
			allocation_value = $3,
			updated_at = $4
		WHERE id = $1
	`
)

func (r *jarRepository) Create(ctx context.Context, j *jar.Jar) error {
	exec := dbExecutor(ctx, r.db)

	_, err := exec.Exec(
		ctx,
		insertJarQuery,
		jarValues(j)...,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf(
				"create jar: %w",
				jar.ErrJarNameExists,
			)
		}

		return fmt.Errorf("create jar: %w", err)
	}

	return nil
}

func (r *jarRepository) List(ctx context.Context) ([]*jar.Jar, error) {
	exec := dbExecutor(ctx, r.db)

	rows, err := exec.Query(ctx, listJarsQuery)
	if err != nil {
		return nil, fmt.Errorf("list jars: %w", err)
	}
	defer rows.Close()

	jars := make([]*jar.Jar, 0)

	for rows.Next() {
		var j jar.Jar

		if err := rows.Scan(jarScanArgs(&j)...); err != nil {
			return nil, fmt.Errorf("scan jar: %w", err)
		}

		jars = append(jars, &j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list jars: %w", err)
	}

	return jars, nil
}

func (r *jarRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*jar.Jar, error) {
	exec := dbExecutor(ctx, r.db)

	var j jar.Jar

	err := exec.QueryRow(
		ctx,
		findJarByIDQuery,
		id,
	).Scan(jarScanArgs(&j)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"find jar: %w",
				jar.ErrJarNotFound,
			)
		}

		return nil, fmt.Errorf("find jar: %w", err)
	}

	return &j, nil
}

func (r *jarRepository) Save(
	ctx context.Context,
	j *jar.Jar,
) error {
	exec := dbExecutor(ctx, r.db)

	tag, err := exec.Exec(
		ctx,
		saveJarQuery,
		jarSaveValues(j)...,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf(
				"save jar: %w",
				jar.ErrJarNameExists,
			)
		}

		return fmt.Errorf("save jar: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"save jar: %w",
			jar.ErrJarNotFound,
		)
	}

	return nil
}

func (r *jarRepository) UpdateAllocations(
	ctx context.Context,
	jars []*jar.Jar,
) error {
	exec := dbExecutor(ctx, r.db)

	for _, j := range jars {
		tag, err := exec.Exec(
			ctx,
			updateJarAllocationQuery,
			j.ID,
			j.AllocationType,
			j.AllocationValue,
			j.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf(
				"update jar allocations: %w",
				err,
			)
		}

		if tag.RowsAffected() == 0 {
			return fmt.Errorf(
				"update jar allocations: %w",
				jar.ErrJarNotFound,
			)
		}
	}

	return nil
}
