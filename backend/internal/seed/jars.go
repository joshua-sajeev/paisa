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
	iconKey        string
	allocationType string
	value          int64
}

var jarDefinitions = []jarSeed{
	{
		name:           "Necessities",
		iconKey:        "jar",
		allocationType: "remainder",
		value:          0,
	},
	{
		name:           "Leisure",
		iconKey:        "jar",
		allocationType: "percentage",
		value:          10,
	},
	{
		name:           "Investment",
		iconKey:        "jar",
		allocationType: "fixed",
		value:          1000000,
	},
	{
		name:           "Donation",
		iconKey:        "jar",
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
				icon_key,
				allocation_type,
				allocation_value,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, FALSE, $6, $6)
			`,
			id,
			def.name,
			def.iconKey,
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
