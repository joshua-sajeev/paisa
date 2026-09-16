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
	DashboardHandler   *handler.DashboardHandler
	AccountHandler     *handler.AccountHandler
	JarHandler         *handler.JarHandler
	TransactionHandler *handler.TransactionHandler
	AuthHandler        *handler.AuthHandler
	GoalHandler        *handler.GoalHandler
	SessionHandler     *handler.SessionHandler
	// Internal dependencies
	logger *slog.Logger
	db     *pgxpool.Pool
	cfg    *config.Config

	// Session
	SessionStore session.SessionStore

	// Repositories
	dashboardRepository   ports.DashboardRepository
	accountRepository     ports.AccountRepository
	jarRepository         ports.JarRepository
	transactionRepository ports.TransactionRepository
	allocationRepository  ports.AllocationRepository
	goalRepository        ports.GoalRepository
	txManager             ports.TxManager

	// Services
	dashboardService   *application.DashboardService
	accountService     *application.AccountService
	jarService         *application.JarService
	transactionService *application.TransactionService
	authService        *application.AuthService
	goalService        *application.GoalService
}

var (
	_ handler.AccountService     = (*application.AccountService)(nil)
	_ handler.JarService         = (*application.JarService)(nil)
	_ handler.TransactionService = (*application.TransactionService)(nil)
)

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
	c.dashboardRepository = postgres.NewDashboardRepository(c.db)
	c.accountRepository = postgres.NewAccountRepository(c.db)
	c.jarRepository = postgres.NewJarRepository(c.db)
	c.transactionRepository = postgres.NewTransactionRepository(c.db)
	c.goalRepository = postgres.NewGoalRepository(c.db)

	c.allocationRepository = postgres.NewAllocationRepository(c.db)

	c.txManager = postgres.NewTxManager(c.db)
}

// initServices creates all service instances with repository dependencies.
func (c *Container) initServices() {
	c.dashboardService = application.NewDashboardService(
		c.dashboardRepository,
		c.logger,
	)
	c.accountService = application.NewAccountService(
		c.accountRepository,
		c.logger,
	)

	c.jarService = application.NewJarService(
		c.jarRepository,
		c.logger,
	)

	c.transactionService = application.NewTransactionService(
		c.transactionRepository,
		c.allocationRepository,
		c.jarRepository,
		c.accountRepository,
		c.txManager,

		c.logger,
	)
	c.authService = application.NewAuthService(
		c.SessionStore,
		c.cfg.AppLock.PINHash,
		c.cfg.SessionTTLMinutes,
	)
	c.goalService = application.NewGoalService(
		c.goalRepository,
		c.logger,
	)
}

// initHandlers creates all handler instances with service dependencies.
func (c *Container) initHandlers() {
	c.DashboardHandler = handler.NewDashboardHandler(
		c.dashboardService,
		c.logger,
	)
	c.AccountHandler = handler.NewAccountHandler(
		c.accountService,
		c.logger,
	)

	c.JarHandler = handler.NewJarHandler(
		c.jarService,
		c.logger,
	)

	c.TransactionHandler = handler.NewTransactionHandler(
		c.transactionService,
		c.logger,
	)

	c.AuthHandler = handler.NewAuthHandler(
		c.authService,
		c.logger,
	)

	c.GoalHandler = handler.NewGoalHandler(
		c.goalService,
		c.logger,
	)

	c.SessionHandler = handler.NewSessionHandler(
		c.cfg.DemoMode,
		c.logger,
	)
}

// Logger returns the application logger.
func (c *Container) Logger() *slog.Logger {
	return c.logger
}

// Close cleans up resources.
func (c *Container) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
