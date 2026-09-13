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

var categories = []string{
	"food",
	"transport",
	"entertainment",
	"groceries",
	"health",
	"transfer",
	"donation",
	"investment",
	"housing",
	"other",
}

type generatedTransaction struct {
	id             uuid.UUID
	name           string
	txType         string
	category       string
	fromAccountID  *uuid.UUID
	toAccountID    *uuid.UUID
	jarID          *uuid.UUID
	amount         int64
	occurredAt     time.Time
	isMasterIncome bool
}

func seedTransactions(
	ctx context.Context,
	db *pgxpool.Pool,
	accounts []uuid.UUID,
	jars []uuid.UUID,
	count int,
) error {
	if len(accounts) == 0 {
		return fmt.Errorf("cannot seed transactions without accounts")
	}

	rng := rand.New(rand.NewSource(42))

	const batchSize = 1_000

	startDate := time.Date(
		2024,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	for offset := 0; offset < count; offset += batchSize {
		end := min(offset+batchSize, count)

		rows := make([][]any, 0, end-offset)
		now := time.Now()

		for i := offset; i < end; i++ {
			tx := generateTransaction(
				rng,
				accounts,
				jars,
				startDate,
			)

			rows = append(rows, []any{
				tx.id,
				tx.name,
				tx.txType,
				tx.category,
				tx.fromAccountID,
				tx.toAccountID,
				tx.jarID,
				tx.amount,
				tx.occurredAt,
				tx.isMasterIncome,
				now,
				now,
			})
		}

		_, err := db.CopyFrom(
			ctx,
			pgx.Identifier{"transactions"},
			[]string{
				"id",
				"name",
				"type",
				"category",
				"from_account_id",
				"to_account_id",
				"jar_id",
				"amount",
				"occurred_at",
				"is_master_income",
				"created_at",
				"updated_at",
			},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return fmt.Errorf(
				"insert transactions %d-%d: %w",
				offset,
				end,
				err,
			)
		}

		if end%100_000 == 0 || end == count {
			fmt.Printf(
				"transactions: %d/%d\n",
				end,
				count,
			)
		}
	}

	return nil
}

func generateTransaction(
	rng *rand.Rand,
	accounts []uuid.UUID,
	jars []uuid.UUID,
	startDate time.Time,
) generatedTransaction {
	from := randomAccount(accounts, rng)
	to := randomAccount(accounts, rng)

	for from == to {
		to = randomAccount(accounts, rng)
	}

	txType := randomTransactionType(rng)
	category := randomCategory(rng)

	amount := int64(100 + rng.Intn(200_000))

	occurredAt := startDate.AddDate(
		0,
		0,
		rng.Intn(365*3),
	)

	var fromID *uuid.UUID
	var toID *uuid.UUID
	var jarID *uuid.UUID

	switch txType {
	case "income":
		toID = &to

	case "expense":
		fromID = &from

		if len(jars) > 0 && rng.Intn(100) < 35 {
			j := jars[rng.Intn(len(jars))]
			jarID = &j
		}

	case "transfer":
		fromID = &from
		toID = &to
		category = "transfer"
	}

	isMasterIncome := txType == "income" && rng.Intn(100) < 20

	return generatedTransaction{
		id:             uuid.New(),
		name:           transactionName(txType, category, rng),
		txType:         txType,
		category:       category,
		fromAccountID:  fromID,
		toAccountID:    toID,
		jarID:          jarID,
		amount:         amount,
		occurredAt:     occurredAt,
		isMasterIncome: isMasterIncome,
	}
}

func randomTransactionType(rng *rand.Rand) string {
	n := rng.Intn(100)

	switch {
	case n < 25:
		return "income"
	case n < 85:
		return "expense"
	default:
		return "transfer"
	}
}

func randomCategory(rng *rand.Rand) string {
	return categories[rng.Intn(len(categories))]
}

func randomAccount(
	accounts []uuid.UUID,
	rng *rand.Rand,
) uuid.UUID {
	return accounts[rng.Intn(len(accounts))]
}

func transactionName(
	txType string,
	category string,
	rng *rand.Rand,
) string {
	names := map[string][]string{
		"food": {
			"Restaurant",
			"Lunch",
			"Dinner",
			"Coffee",
		},
		"transport": {
			"Fuel",
			"Uber",
			"Bus",
			"Metro",
		},
		"entertainment": {
			"Movie",
			"Game",
			"Subscription",
			"Concert",
		},
		"groceries": {
			"Supermarket",
			"Groceries",
			"Monthly Groceries",
		},
		"health": {
			"Pharmacy",
			"Doctor",
			"Medicine",
		},
		"donation": {
			"Donation",
			"Charity",
		},
		"investment": {
			"Investment",
			"Mutual Fund",
			"Stocks",
		},
		"housing": {
			"Rent",
			"Electricity",
			"Internet",
		},
		"transfer": {
			"Account Transfer",
		},
		"other": {
			"Miscellaneous",
			"Other Expense",
		},
	}

	options, ok := names[category]
	if !ok {
		return fmt.Sprintf("%s transaction", txType)
	}

	return options[rng.Intn(len(options))]
}
