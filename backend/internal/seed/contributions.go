package seed

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const contributionsPerGoal = 12

func seedContributions(
	ctx context.Context,
	db *pgxpool.Pool,
	goals []uuid.UUID,
) error {
	rng := rand.New(rand.NewSource(42))

	rows := make([][]any, 0, len(goals)*contributionsPerGoal)

	startDate := time.Date(
		2025,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	now := time.Now()

	for _, goalID := range goals {
		for i := range contributionsPerGoal {
			id := uuid.New()

			amount := int64(
				5_000 + rng.Intn(45_000),
			)

			occurredAt := startDate.AddDate(
				0,
				i,
				rng.Intn(28),
			)

			rows = append(rows, []any{
				id,
				goalID,
				amount,
				occurredAt,
				now,
				now,
			})
		}
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
