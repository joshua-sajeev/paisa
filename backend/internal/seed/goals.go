package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type goalSeed struct {
	name                string
	target              int64
	initialContribution int64
	deadline            *time.Time
}

var goalDefinitions = []goalSeed{
	{
		name:                "Laptop",
		target:              7_500_000,
		initialContribution: 100_000,
		deadline:            date(2027, 3, 31),
	},
	{
		name:                "Emergency Fund",
		target:              10_000_000,
		initialContribution: 500_000,
		deadline:            date(2027, 6, 30),
	},
}

func seedGoals(
	ctx context.Context,
	db *pgxpool.Pool,
) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(goalDefinitions))
	now := time.Now()

	for i, goal := range goalDefinitions {
		id := uuid.New()

		_, err := db.Exec(
			ctx,
			`
			INSERT INTO goals (
				id,
				name,
				target,
				deadline,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, FALSE, $5, $5)
			`,
			id,
			goal.name,
			goal.target,
			goal.deadline,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"insert goal %d: %w",
				i+1,
				err,
			)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func date(year int, month time.Month, day int) *time.Time {
	value := time.Date(
		year,
		month,
		day,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	return &value
}
