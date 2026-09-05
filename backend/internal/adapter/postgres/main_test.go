package postgres_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joshu-sajeev/paisa/internal/adapter/postgres"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	db          *pgxpool.Pool
	accountRepo ports.AccountRepository
	ctx         = context.Background()
)

// truncateAccountsTable clears all data from accounts table and related tables
func truncateAccountsTable(t testing.TB, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	_, err := db.Exec(ctx, "TRUNCATE TABLE accounts CASCADE")
	if err != nil {
		t.Fatalf("truncate accounts table: %v", err)
	}
}

func TestMain(m *testing.M) {
	container, err := tcpostgres.Run(
		ctx,
		"postgres:18",
		tcpostgres.WithDatabase("paisa_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminate postgres container: %v", err)
		}
	}()

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("connection string: %v", err)
	}

	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("open sql db: %v", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	if err := goose.Up(
		sqlDB,
		filepath.Join("..", "..", "..", "migrations"),
	); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	db, err = pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("create pgx pool: %v", err)
	}
	defer db.Close()

	accountRepo = postgres.NewAccountRepository(db)
	os.Exit(m.Run())
}
