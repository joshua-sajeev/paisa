package seed

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
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

	if err := refreshAccountBalances(ctx, db); err != nil {
		return fmt.Errorf("refresh account balances: %w", err)
	}

	return nil
}

// seedDemoTransactions creates a small, repeatable dataset around the current
// and previous calendar months for the default development seed.
func seedDemoTransactions(
	ctx context.Context,
	db *pgxpool.Pool,
	accounts []uuid.UUID,
	jars []uuid.UUID,
) error {
	if len(accounts) < 4 || len(jars) < 4 {
		return fmt.Errorf("demo seed requires four accounts and four jars")
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	previousMonthStart := monthStart.AddDate(0, -1, 0)
	accountsByName := map[string]uuid.UUID{}
	for i, def := range accountDefinitions {
		accountsByName[def.name] = accounts[i]
	}
	jarIDsByName := map[string]uuid.UUID{}
	for i, def := range jarDefinitions {
		jarIDsByName[def.name] = jars[i]
	}

	hdfc := accountsByName["HDFC"]
	fbi := accountsByName["FBI"]
	wallet := accountsByName["Wallet"]
	var transactions []generatedTransaction
	var masterIncomeIDs []uuid.UUID
	type incomeAllocation struct {
		transactionID uuid.UUID
		amount        int64
	}
	var appaIncomeAllocations []incomeAllocation
	add := func(name, txType, category string, from, to *uuid.UUID, amount int64, occurredAt time.Time, master bool, jarID *uuid.UUID) uuid.UUID {
		tx := generatedTransaction{
			id: uuid.New(), name: name, txType: txType, category: category,
			fromAccountID: from, toAccountID: to, jarID: jarID,
			amount: amount, occurredAt: occurredAt, isMasterIncome: master,
		}
		transactions = append(transactions, tx)
		if master {
			masterIncomeIDs = append(masterIncomeIDs, tx.id)
		}
		return tx.id
	}
	timeRNG := rand.New(rand.NewSource(20261012))
	transactionTime := func(month time.Time, date int, name string) time.Time {
		hour := 12
		switch {
		case name == "Tea":
			hour = 16 + timeRNG.Intn(4)
		case name == "Breakfast":
			hour = 7 + timeRNG.Intn(2)
		case name == "BMTC":
			if timeRNG.Intn(2) == 0 {
				hour = 9
			} else {
				hour = 17
			}
		case name == "Church":
			hour = 11
		case name == "Rent":
			hour = 19
		case name == "Salary":
			hour = 9
		case name == "SIP":
			hour = 10
		case name == "Groceries":
			hour = 17 + timeRNG.Intn(4)
		case name == "Cab" || name == "Uber":
			hour = 8 + timeRNG.Intn(14)
		case strings.HasPrefix(name, "Bus to "):
			hour = 8 + timeRNG.Intn(3)
		case name == "Dentist" || name == "Hair Cut":
			hour = 10 + timeRNG.Intn(7)
		case name == "Falooda":
			hour = 15 + timeRNG.Intn(6)
		case name == "Appa":
			hour = 13
		case name == "Brother":
			hour = 14
		}
		return time.Date(month.Year(), month.Month(), date, hour, timeRNG.Intn(60), 0, 0, month.Location())
	}
	necessities := jarIDsByName["Necessities"]
	leisure := jarIDsByName["Leisure"]
	donations := jarIDsByName["Donations"]
	investments := jarIDsByName["Investments"]
	addExpense := func(month time.Time, date int, name string, amountRupees int64, category string, accountID, jarID uuid.UUID) {
		add(name, "expense", category, &accountID, nil, amountRupees*100, transactionTime(month, date, name), false, &jarID)
	}

	for _, month := range []time.Time{previousMonthStart, monthStart} {
		add("Salary", "income", "other", nil, &hdfc, 2_500_000, transactionTime(month, 1, "Salary"), true, nil)
		add("Rent", "expense", "housing", &hdfc, nil, 1_200_000, transactionTime(month, 1, "Rent"), false, &necessities)
		addExpense(month, 20, "SIP", 1_000, "investment", fbi, investments)
	}
	add("HDFC to FBI", "transfer", "transfer", &hdfc, &fbi, 500_000, transactionTime(previousMonthStart, 15, "HDFC to FBI"), false, nil)
	add("Brother", "transfer", "transfer", &hdfc, &fbi, 50_000, transactionTime(previousMonthStart, 15, "Brother"), false, nil)

	// Previous month transactions.
	appaIncomeAllocations = append(appaIncomeAllocations, incomeAllocation{
		transactionID: add("Appa", "income", "other", nil, &wallet, 50_000, transactionTime(previousMonthStart, 28, "Appa"), false, &necessities),
		amount:        50_000,
	})
	addExpense(previousMonthStart, 7, "Church", 100, "donation", hdfc, donations)
	addExpense(previousMonthStart, 7, "Tea", 10, "food", hdfc, leisure)
	addExpense(previousMonthStart, 7, "Groceries", 230, "groceries", hdfc, necessities)
	addExpense(previousMonthStart, 8, "BMTC", 12, "transport", hdfc, necessities)
	addExpense(previousMonthStart, 8, "BMTC", 23, "transport", hdfc, necessities)
	addExpense(previousMonthStart, 10, "Cab", 120, "transport", hdfc, leisure)
	addExpense(previousMonthStart, 13, "Tea", 10, "food", hdfc, leisure)
	addExpense(previousMonthStart, 13, "Cab", 73, "transport", hdfc, leisure)
	addExpense(previousMonthStart, 15, "BMTC", 18, "transport", hdfc, leisure)

	// Current month transactions.
	appaIncomeAllocations = append(appaIncomeAllocations, incomeAllocation{
		transactionID: add("Appa", "income", "other", nil, &wallet, 50_000, transactionTime(monthStart, 7, "Appa"), false, &necessities),
		amount:        50_000,
	})
	addExpense(monthStart, 1, "Falooda", 150, "food", hdfc, leisure)
	addExpense(monthStart, 13, "BMTC", 18, "transport", hdfc, leisure)
	addExpense(monthStart, 13, "Tea", 20, "food", hdfc, leisure)
	addExpense(monthStart, 13, "Groceries", 180, "groceries", hdfc, leisure)
	addExpense(monthStart, 18, "Bus to Alakode", 1_500, "transport", hdfc, necessities)
	addExpense(monthStart, 18, "Bus to Bangalore", 700, "transport", hdfc, necessities)
	addExpense(monthStart, 20, "BMTC", 18, "transport", hdfc, leisure)
	addExpense(monthStart, 23, "Hair Cut", 50, "other", hdfc, necessities)
	addExpense(monthStart, 23, "Dentist", 500, "health", hdfc, necessities)
	addExpense(monthStart, 25, "Groceries", 150, "groceries", hdfc, necessities)
	addExpense(monthStart, 28, "Church", 200, "donation", hdfc, donations)
	addExpense(monthStart, 13, "Uber", 200, "transport", hdfc, leisure)
	if now.Day() >= 25 {
		appaIncomeAllocations = append(appaIncomeAllocations, incomeAllocation{
			transactionID: add("Appa", "income", "other", nil, &hdfc, 100_000, transactionTime(monthStart, 25, "Appa"), false, &necessities),
			amount:        100_000,
		})
	}

	// Keep the sample cap at thirty without adding any current-month records
	// dated after today. Previous-month transactions remain in the seed.
	currentMonthTransactions := transactions[:0]
	for _, tx := range transactions {
		if tx.occurredAt.Year() == now.Year() &&
			tx.occurredAt.Month() == now.Month() &&
			tx.occurredAt.Day() > now.Day() {
			continue
		}
		currentMonthTransactions = append(currentMonthTransactions, tx)
	}
	transactions = currentMonthTransactions

	rows := make([][]any, 0, len(transactions))
	for _, tx := range transactions {
		rows = append(rows, []any{
			tx.id, tx.name, tx.txType, tx.category, tx.fromAccountID,
			tx.toAccountID, tx.jarID, tx.amount, tx.occurredAt,
			tx.isMasterIncome, now, now,
		})
	}
	if _, err := db.CopyFrom(ctx, pgx.Identifier{"transactions"}, []string{
		"id", "name", "type", "category", "from_account_id", "to_account_id",
		"jar_id", "amount", "occurred_at", "is_master_income", "created_at", "updated_at",
	}, pgx.CopyFromRows(rows)); err != nil {
		return fmt.Errorf("insert demo transactions: %w", err)
	}

	// Per 25,000 salary: 14,000 remainder, 2,500 each at 10%, and 6,000 fixed.
	allocations := make([][]any, 0, len(masterIncomeIDs)*4)
	for _, transactionID := range masterIncomeIDs {
		for _, allocation := range []struct {
			jarID  uuid.UUID
			amount int64
		}{
			{jarIDsByName["Necessities"], 1_400_000},
			{jarIDsByName["Leisure"], 250_000},
			{jarIDsByName["Donations"], 250_000},
			{jarIDsByName["Investments"], 600_000},
		} {
			allocations = append(allocations, []any{uuid.New(), transactionID, allocation.jarID, allocation.amount, now})
		}
	}
	for _, appaIncome := range appaIncomeAllocations {
		allocations = append(allocations, []any{
			uuid.New(), appaIncome.transactionID, necessities, appaIncome.amount, now,
		})
	}
	if _, err := db.CopyFrom(ctx, pgx.Identifier{"jar_allocations"}, []string{
		"id", "transaction_id", "jar_id", "amount", "created_at",
	}, pgx.CopyFromRows(allocations)); err != nil {
		return fmt.Errorf("insert demo income allocations: %w", err)
	}
	if err := refreshAccountBalances(ctx, db); err != nil {
		return fmt.Errorf("refresh account balances: %w", err)
	}
	fmt.Printf("transactions: %d/%d\n", len(transactions), transactionCount)
	return nil
}

func refreshAccountBalances(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(
		ctx,
		`
		WITH account_deltas AS (
			SELECT
				account_id,
				SUM(delta) AS balance
			FROM (
				SELECT
					to_account_id AS account_id,
					amount AS delta
				FROM transactions
				WHERE to_account_id IS NOT NULL

				UNION ALL

				SELECT
					from_account_id AS account_id,
					-amount AS delta
				FROM transactions
				WHERE from_account_id IS NOT NULL
			) deltas
			GROUP BY account_id
		)
		UPDATE accounts a
		SET
			balance = a.balance + account_deltas.balance,
			updated_at = NOW()
		FROM account_deltas
		WHERE a.id = account_deltas.account_id
		`,
	)
	if err != nil {
		return err
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
	).Add(time.Duration(rng.Intn(24*60)) * time.Minute)

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
