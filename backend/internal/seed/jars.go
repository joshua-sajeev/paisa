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
	rollover       bool
}

var jarDefinitions = []jarSeed{
	{
		name:           "Necessities",
		iconKey:        "necessities",
		allocationType: "remainder",
		value:          0,
		rollover:       true,
	},
	{
		name:           "Leisure",
		iconKey:        "leisure",
		allocationType: "percentage",
		value:          10,
		rollover:       false,
	},
	{
		name:           "Investments",
		iconKey:        "investment",
		allocationType: "fixed",
		value:          600000,
		rollover:       true,
	},
	{
		name:           "Donations",
		iconKey:        "giving",
		allocationType: "percentage",
		value:          10,
		rollover:       true,
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
				rollover,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, FALSE, $7, $7)
			`,
			id,
			def.name,
			def.iconKey,
			def.allocationType,
			def.value,
			def.rollover,
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
