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

// Container holds the application dependencies required by the HTTP layer.
type Container struct {
	// HTTP Handlers
	AccountHandler *handler.AccountHandler
	AuthHandler    *handler.AuthHandler

	// Internal dependencies
	logger *slog.Logger
	db     *pgxpool.Pool

	// Repositories
	accountRepository ports.AccountRepository

	// Services
	accountService *application.AccountService

	// Session store
	sessionStore session.SessionStore
}

var _ handler.AccountService = (*application.AccountService)(nil)

// New creates and initializes the dependency container.
func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
	}

	if err := c.initDatabase(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	c.initRepositories()
	c.initServices()

	// Initialize session store unless demo mode is enabled.
	if !cfg.AppLock.DemoMode {
		store, err := session.NewRedisSessionStore(cfg.SessionStoreURL)
		if err != nil {
			return nil, fmt.Errorf("failed to init session store: %w", err)
		}

		c.sessionStore = store
	}

	c.initHandlers(cfg)

	return c, nil
}

// initDatabase establishes a connection pool to PostgreSQL.
func (c *Container) initDatabase(
	ctx context.Context,
	cfg *config.Config,
) error {
	dbURL := cfg.Database.GetDatabaseURL()

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
}

// initHandlers creates all handler instances with service dependencies.
func (c *Container) initHandlers(cfg *config.Config) {
	c.AccountHandler = handler.NewAccountHandler(
		c.accountService,
		c.logger,
	)

	// Auth handler. In demo mode the session store is intentionally nil.
	var store session.SessionStore
	if !cfg.AppLock.DemoMode {
		store = c.sessionStore
	}

	c.AuthHandler = handler.NewAuthHandler(
		cfg,
		store,
		c.logger,
	)
}

// Logger returns the application logger.
func (c *Container) Logger() *slog.Logger {
	return c.logger
}

// SessionStore returns the application session store.
// It is nil when demo mode is enabled.
func (c *Container) SessionStore() session.SessionStore {
	return c.sessionStore
}

// Close cleans up resources.
func (c *Container) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
