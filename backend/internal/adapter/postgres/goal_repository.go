package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// goalRepository implements the GoalRepository port using PostgreSQL.
type goalRepository struct {
	db *pgxpool.Pool
}

// NewGoalRepository creates and returns the goal repository.
func NewGoalRepository(db *pgxpool.Pool) ports.GoalRepository {
	return &goalRepository{db: db}
}

func goalScanArgs(g *goal.Goal) []any {
	return []any{
		&g.ID,
		&g.Name,
		&g.Target,
		&g.Deadline,
		&g.IsArchived,
		&g.CreatedAt,
		&g.UpdatedAt,
	}
}

const (
	insertGoalQuery = `
		INSERT INTO goals (id, name, target, deadline, is_archived, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	listGoalsQuery = `
		SELECT id, name, target, deadline, is_archived, created_at, updated_at
		FROM goals
		ORDER BY created_at DESC, id DESC
	`
	findGoalByIDQuery = `
		SELECT id, name, target, deadline, is_archived, created_at, updated_at
		FROM goals
		WHERE id = $1
	`
	saveGoalQuery = `
		UPDATE goals
		SET name = $2, target = $3, deadline = $4, is_archived = $5, updated_at = $6
		WHERE id = $1
	`
	insertContributionQuery = `
		INSERT INTO contributions (id, goal_id, amount, occurred_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	listContributionsQuery = `
		SELECT id, goal_id, amount, occurred_at, created_at, updated_at
		FROM contributions
		WHERE goal_id = $1
		ORDER BY occurred_at DESC, created_at DESC
	`
	deleteGoalQuery = `
		DELETE FROM goals
		WHERE id = $1
	`
	deleteContributionsByGoalQuery = `
		DELETE FROM contributions
		WHERE goal_id = $1
	`
)

func (r *goalRepository) Create(ctx context.Context, g *goal.Goal) error {
	exec := dbExecutor(ctx, r.db)
	_, err := exec.Exec(ctx, insertGoalQuery, g.ID, g.Name, g.Target, g.Deadline, g.IsArchived, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("create goal: %w", goal.ErrGoalNameExists)
		}
		return fmt.Errorf("create goal: %w", err)
	}
	return nil
}

func (r *goalRepository) List(ctx context.Context) ([]*goal.Goal, error) {
	exec := dbExecutor(ctx, r.db)
	rows, err := exec.Query(ctx, listGoalsQuery)
	if err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}
	defer rows.Close()

	goals := make([]*goal.Goal, 0)
	for rows.Next() {
		var g goal.Goal
		if err := rows.Scan(goalScanArgs(&g)...); err != nil {
			return nil, fmt.Errorf("scan goal: %w", err)
		}
		goals = append(goals, &g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}
	return goals, nil
}

func (r *goalRepository) FindByID(ctx context.Context, id uuid.UUID) (*goal.Goal, error) {
	exec := dbExecutor(ctx, r.db)
	var g goal.Goal
	err := exec.QueryRow(ctx, findGoalByIDQuery, id).Scan(goalScanArgs(&g)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("find goal: %w", goal.ErrGoalNotFound)
		}
		return nil, fmt.Errorf("find goal: %w", err)
	}
	return &g, nil
}

func (r *goalRepository) Save(ctx context.Context, g *goal.Goal) error {
	exec := dbExecutor(ctx, r.db)
	tag, err := exec.Exec(ctx, saveGoalQuery, g.ID, g.Name, g.Target, g.Deadline, g.IsArchived, g.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("save goal: %w", goal.ErrGoalNameExists)
		}
		return fmt.Errorf("save goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("save goal: %w", goal.ErrGoalNotFound)
	}
	return nil
}

func (r *goalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return r.deleteWithExec(ctx, tx, id)
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.deleteWithExec(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *goalRepository) deleteWithExec(ctx context.Context, exec dbExec, id uuid.UUID) error {
	if _, err := exec.Exec(ctx, deleteContributionsByGoalQuery, id); err != nil {
		return fmt.Errorf("delete contributions: %w", err)
	}

	tag, err := exec.Exec(ctx, deleteGoalQuery, id)
	if err != nil {
		return fmt.Errorf("delete goal: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete goal: %w", goal.ErrGoalNotFound)
	}

	return nil
}

func (r *goalRepository) AddContribution(ctx context.Context, c *goal.Contribution) error {
	exec := dbExecutor(ctx, r.db)
	_, err := exec.Exec(ctx, insertContributionQuery, c.ID, c.GoalID, c.Amount, c.OccurredAt, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("add contribution: %w", err)
	}
	return nil
}

func (r *goalRepository) ListContributions(ctx context.Context, goalID uuid.UUID) ([]*goal.Contribution, error) {
	exec := dbExecutor(ctx, r.db)
	rows, err := exec.Query(ctx, listContributionsQuery, goalID)
	if err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}
	defer rows.Close()

	contributions := make([]*goal.Contribution, 0)
	for rows.Next() {
		var c goal.Contribution
		err := rows.Scan(
			&c.ID,
			&c.GoalID,
			&c.Amount,
			&c.OccurredAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan contribution: %w", err)
		}
		contributions = append(contributions, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list contributions: %w", err)
	}
	return contributions, nil
}
