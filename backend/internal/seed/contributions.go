package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func seedContributions(
	ctx context.Context,
	db *pgxpool.Pool,
	goals []uuid.UUID,
) error {
	if len(goals) != len(goalDefinitions) {
		return fmt.Errorf("got %d goals for %d goal definitions", len(goals), len(goalDefinitions))
	}
	rows := make([][]any, 0, len(goals))
	now := time.Now()
	for i, goalID := range goals {
		rows = append(rows, []any{
			uuid.New(),
			goalID,
			goalDefinitions[i].initialContribution,
			now,
			now,
			now,
		})
	}

	_, err := db.CopyFrom(
		ctx,
		pgx.Identifier{"contributions"},
		[]string{
			"id",
			"goal_id",
			"amount",
			"occurred_at",
			"created_at",
			"updated_at",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy contributions: %w", err)
	}

	return nil
}
