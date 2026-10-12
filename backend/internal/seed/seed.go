// Package seed provides seeding data for Paisa.
package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const transactionCount = 30

func Run(ctx context.Context, db *pgxpool.Pool) error {
	return run(ctx, db, transactionCount, true)
}

func RunWithCount(ctx context.Context, db *pgxpool.Pool, count int) error {
	return run(ctx, db, count, false)
}

func run(ctx context.Context, db *pgxpool.Pool, count int, demo bool) error {
	transactionCountLabel := "%d transactions"
	if demo {
		transactionCountLabel = "up to %d transactions"
	}
	log.Printf(
		"seeding dataset: %d accounts, %d jars, %d goals, "+transactionCountLabel,
		len(accountDefinitions),
		len(jarDefinitions),
		len(goalDefinitions),
		count,
	)

	if err := clearSeedData(ctx, db); err != nil {
		return fmt.Errorf("clear seed data: %w", err)
	}

	accounts, err := seedAccounts(ctx, db)
	if err != nil {
		return fmt.Errorf("seed accounts: %w", err)
	}

	jars, err := seedJars(ctx, db)
	if err != nil {
		return fmt.Errorf("seed jars: %w", err)
	}

	goals, err := seedGoals(ctx, db)
	if err != nil {
		return fmt.Errorf("seed goals: %w", err)
	}

	if err := seedContributions(ctx, db, goals); err != nil {
		return fmt.Errorf("seed contributions: %w", err)
	}

	var transactionErr error
	if demo {
		transactionErr = seedDemoTransactions(ctx, db, accounts, jars)
	} else {
		transactionErr = seedTransactions(ctx, db, accounts, jars, count)
	}
	if transactionErr != nil {
		return fmt.Errorf("seed transactions: %w", transactionErr)
	}

	return nil
}
