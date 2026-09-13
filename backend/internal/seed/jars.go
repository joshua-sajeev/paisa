package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type jarSeed struct {
	name           string
	allocationType string
	value          int64
}

var jarDefinitions = []jarSeed{
	{
		name:           "Necessities",
		allocationType: "remainder",
		value:          0,
	},
	{
		name:           "Leisure",
		allocationType: "percentage",
		value:          10,
	},
	{
		name:           "Investment",
		allocationType: "fixed",
		value:          1000000,
	},
	{
		name:           "Donation",
		allocationType: "percentage",
		value:          10,
	},
}

func seedJars(
	ctx context.Context,
	db *pgxpool.Pool,
) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(jarDefinitions))
	now := time.Now()

	for i, def := range jarDefinitions {
		id := uuid.New()

		_, err := db.Exec(
			ctx,
			`
			INSERT INTO jars (
				id,
				name,
				allocation_type,
				allocation_value,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, FALSE, $5, $5)
			`,
			id,
			def.name,
			def.allocationType,
			def.value,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"insert jar %d: %w",
				i+1,
				err,
			)
		}

		ids = append(ids, id)
	}

	return ids, nil
}
