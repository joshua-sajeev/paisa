package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/config"
	"github.com/joshu-sajeev/paisa/internal/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.Database.ConnectionURL())
	if err != nil {
		log.Fatalf("create database pool: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	if err := seed.Run(ctx, db); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	log.Println("seeding completed successfully")
}
