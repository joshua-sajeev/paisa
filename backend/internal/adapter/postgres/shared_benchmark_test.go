package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/seed"
)

func setupBenchmark(b *testing.B, count int) (uuid.UUID, uuid.UUID, []uuid.UUID) {
	b.Helper()

	if err := seed.RunWithCount(ctx, db, count); err != nil {
		b.Fatalf("failed to seed: %v", err)
	}

	// Fetch accounts from DB to return active IDs
	var accIDs []uuid.UUID
	rows, err := db.Query(ctx, "SELECT id FROM accounts ORDER BY name")
	if err != nil {
		b.Fatalf("failed to query accounts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			b.Fatalf("failed to scan account id: %v", err)
		}
		accIDs = append(accIDs, id)
	}
	if len(accIDs) < 2 {
		b.Fatalf("expected at least 2 accounts, got %d", len(accIDs))
	}

	// Fetch transactions from DB to return active IDs
	var txIDs []uuid.UUID
	txRows, err := db.Query(ctx, "SELECT id FROM transactions LIMIT $1", count)
	if err != nil {
		b.Fatalf("failed to query transactions: %v", err)
	}
	defer txRows.Close()

	for txRows.Next() {
		var id uuid.UUID
		if err := txRows.Scan(&id); err != nil {
			b.Fatalf("failed to scan transaction id: %v", err)
		}
		txIDs = append(txIDs, id)
	}

	return accIDs[0], accIDs[1], txIDs
}
