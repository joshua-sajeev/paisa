// Package seed provides seeding data for Paisa.
package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const transactionCount = 10_000

func Run(ctx context.Context, db *pgxpool.Pool) error {
	return RunWithCount(ctx, db, transactionCount)
}

func RunWithCount(ctx context.Context, db *pgxpool.Pool, count int) error {
	log.Printf(
		"seeding dataset: %d accounts, %d jars, %d goals, %d transactions",
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

	if err := seedTransactions(
		ctx,
		db,
		accounts,
		jars,
		count,
	); err != nil {
		return fmt.Errorf("seed transactions: %w", err)
	}

	return nil
}
