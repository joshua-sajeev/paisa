package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var accountDefinitions = []struct {
	name           string
	iconKey        string
	openingBalance int64
	isPrimary      bool
}{
	{
		name:           "FBI",
		iconKey:        "fbi",
		openingBalance: 200_000,
	},
	{
		name:           "SBI",
		iconKey:        "sbi",
		openingBalance: 1_050_000,
	},
	{
		name:           "HDFC",
		iconKey:        "hdfc",
		openingBalance: 500_000,
		isPrimary:      true,
	},
	{
		name:           "Wallet",
		iconKey:        "cash",
		openingBalance: 70_000,
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
				balance,
				is_primary,
				is_archived,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, FALSE, $6, $6)
			`,
			id,
			account.name,
			account.iconKey,
			account.openingBalance,
			account.isPrimary,
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
