package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var accountDefinitions = []string{
	"HDFC",
	"ICICI",
	"HSBC",
	"SBI",
	"FBI",
	"Cash",
}

func seedAccounts(
	ctx context.Context,
	db *pgxpool.Pool,
) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(accountDefinitions))
	now := time.Now()

	for i, name := range accountDefinitions {
		id := uuid.New()

		_, err := db.Exec(
			ctx,
			`
			INSERT INTO accounts (
				id,
				name,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, FALSE, $3, $3)
			`,
			id,
			name,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"insert account %d: %w",
				i+1,
				err,
			)
		}

		ids = append(ids, id)
	}

	return ids, nil
}
