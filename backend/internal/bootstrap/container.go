// Package bootstrap creates and wires all application dependencies
package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/adapter/postgres"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/config"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/joshu-sajeev/paisa/internal/session"
)

type Container struct {
	// HTTP Handlers
	AccountHandler *handler.AccountHandler
	AuthHandler    *handler.AuthHandler

	// Internal dependencies
	logger *slog.Logger
	db     *pgxpool.Pool
	cfg    *config.Config

	// Session
	SessionStore session.SessionStore

	// Repositories
	accountRepository ports.AccountRepository

	// Services
	accountService *application.AccountService
	authService    *application.AuthService
}

var _ handler.AccountService = (*application.AccountService)(nil)

// New creates and initializes the dependency container
func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	c := &Container{
		cfg: cfg,
		logger: slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		)),
		SessionStore: session.NewInMemoryStore(),
	}

	if err := c.initDatabase(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	c.initRepositories()
	c.initServices()
	c.initHandlers()

	return c, nil
}

// initDatabase establishes a connection pool to PostgreSQL.
func (c *Container) initDatabase(
	ctx context.Context,
	cfg *config.Config,
) error {
	dbURL := cfg.Database.ConnectionURL()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.db = pool
	return nil
}

// initRepositories creates all repository instances.
func (c *Container) initRepositories() {
	c.accountRepository = postgres.NewAccountRepository(c.db)
}

// initServices creates all service instances with repository dependencies.
func (c *Container) initServices() {
	c.accountService = application.NewAccountService(
		c.accountRepository,
		c.logger,
	)

	c.authService = application.NewAuthService(
		c.SessionStore,
		c.cfg.AppLock.PINHash,
		c.cfg.SessionTTLMinutes,
	)
}

// initHandlers creates all handler instances with service dependencies.
func (c *Container) initHandlers() {
	c.AccountHandler = handler.NewAccountHandler(
		c.accountService,
		c.logger,
	)

	c.AuthHandler = handler.NewAuthHandler(
		c.authService,
		c.logger,
	)
}

func (c *Container) Logger() *slog.Logger {
	return c.logger
}

// Close cleans up resources.
func (c *Container) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
