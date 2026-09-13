package seed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func clearSeedData(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(
		ctx,
		`
		TRUNCATE TABLE
			contributions,
			jar_allocations,
			transactions,
			templates,
			goals,
			jars,
			accounts
		RESTART IDENTITY CASCADE
		`,
	)
	if err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}

	return nil
}
