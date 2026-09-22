package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var accountDefinitions = []struct {
	name    string
	iconKey string
}{
	{
		name:    "HDFC",
		iconKey: "hdfc",
	},
	{
		name:    "ICICI",
		iconKey: "icici",
	},
	{
		name:    "HSBC",
		iconKey: "hsbc",
	},
	{
		name:    "SBI",
		iconKey: "sbi",
	},
	{
		name:    "FBI",
		iconKey: "bank",
	},
	{
		name:    "Cash",
		iconKey: "cash",
	},
}

func seedAccounts(
	ctx context.Context,
	db *pgxpool.Pool,
) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(accountDefinitions))
	now := time.Now()

	for i, account := range accountDefinitions {
		id := uuid.New()

		_, err := db.Exec(
			ctx,
			`
			INSERT INTO accounts (
				id,
				name,
				icon_key,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, FALSE, $4, $4)
			`,
			id,
			account.name,
			account.iconKey,
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
