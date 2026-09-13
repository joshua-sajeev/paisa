package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type goalSeed struct {
	name     string
	target   int64
	deadline *time.Time
}

var goalDefinitions = []goalSeed{
	{
		name:     "Emergency Fund",
		target:   5000000,
		deadline: date(2027, 6, 30),
	},
	{
		name:     "New Laptop",
		target:   10000000,
		deadline: date(2027, 3, 31),
	},
	{
		name:     "Travel Fund",
		target:   7500000,
		deadline: date(2027, 12, 31),
	},
	{
		name:     "Bike Fund",
		target:   15000000,
		deadline: date(2028, 6, 30),
	},
	{
		name:     "Education",
		target:   5000000,
		deadline: date(2027, 5, 31),
	},
	{
		name:     "Home Setup",
		target:   10000000,
		deadline: date(2028, 3, 31),
	},
	{
		name:     "Investment",
		target:   20000000,
		deadline: date(2029, 12, 31),
	},
	{
		name:     "Wedding",
		target:   50000000,
		deadline: date(2030, 12, 31),
	},
	{
		name:     "Car Fund",
		target:   50000000,
		deadline: date(2030, 6, 30),
	},
	{
		name:     "Misc Savings",
		target:   2500000,
		deadline: date(2027, 12, 31),
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
