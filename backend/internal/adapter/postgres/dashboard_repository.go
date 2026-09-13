package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

type dashboardRepository struct {
	db *pgxpool.Pool
}

// NewDashboardRepository creates a new dashboard repository.
func NewDashboardRepository(db *pgxpool.Pool) ports.DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}

var _ ports.DashboardRepository = (*dashboardRepository)(nil)

// GetTotalBalance returns current total balance across all accounts.
func (r *dashboardRepository) GetTotalBalance(ctx context.Context) (int64, error) {
	db := dbExecutor(ctx, r.db)

	const query = `
		SELECT COALESCE(SUM(balance), 0)
		FROM accounts
		WHERE is_archived = FALSE
	`

	var balance int64
	if err := db.QueryRow(ctx, query).Scan(&balance); err != nil {
		return 0, err
	}

	return balance, nil
}

// GetMonthlySummary returns current month's income, expense, savings, and rate.
func (r *dashboardRepository) GetMonthlySummary(ctx context.Context) (*ports.DashboardSummary, error) {
	db := dbExecutor(ctx, r.db)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	const query = `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0)
		FROM transactions
		WHERE occurred_at >= $1
		  AND occurred_at < $2
	`

	var income, expense int64
	if err := db.QueryRow(ctx, query, monthStart, monthEnd).Scan(&income, &expense); err != nil {
		return nil, err
	}

	savings := income - expense
	savingsRate := 0.0

	if income > 0 {
		savingsRate = float64(savings) / float64(income) * 100
	}

	return &ports.DashboardSummary{
		TotalBalance:       0,
		MonthlyIncome:      income,
		MonthlyExpense:     expense,
		MonthlySavings:     savings,
		MonthlySavingsRate: savingsRate,
	}, nil
}

// GetAccountBalances returns all non-archived accounts with balances.
func (r *dashboardRepository) GetAccountBalances(ctx context.Context) ([]*ports.AccountBalance, error) {
	db := dbExecutor(ctx, r.db)

	const query = `
		SELECT
			id,
			name,
			balance
		FROM accounts
		WHERE is_archived = FALSE
		ORDER BY created_at, id
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*ports.AccountBalance

	for rows.Next() {
		var account ports.AccountBalance

		if err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Balance,
		); err != nil {
			return nil, err
		}

		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

// GetJarSummaries returns all active jars with allocation/spending state.
func (r *dashboardRepository) GetJarSummaries(ctx context.Context) ([]*ports.JarSummary, error) {
	db := dbExecutor(ctx, r.db)

	const query = `
		WITH active_jars AS (
			SELECT
				id,
				name,
				allocation_type,
				allocation_value,
				created_at
			FROM jars
			WHERE is_archived = FALSE
		),
		allocation_totals AS (
			SELECT
				jar_id,
				SUM(amount) AS allocated
			FROM jar_allocations
			GROUP BY jar_id
		),
		expense_totals AS (
			SELECT
				jar_id,
				SUM(amount) AS used
			FROM transactions
			WHERE type = 'expense'
			  AND jar_id IS NOT NULL
			GROUP BY jar_id
		)
		SELECT
			j.id,
			j.name,
			j.allocation_type,
			j.allocation_value,
			COALESCE(a.allocated, 0) AS allocated,
			COALESCE(e.used, 0) AS used,
			GREATEST(
				COALESCE(a.allocated, 0) - COALESCE(e.used, 0),
				0
			) AS available,
			CASE
				WHEN COALESCE(a.allocated, 0) = 0 THEN 0.0
				ELSE (
					COALESCE(e.used, 0)::FLOAT /
					COALESCE(a.allocated, 0) * 100
				)
			END AS used_percentage,
			CASE
				WHEN COALESCE(a.allocated, 0) = 0 THEN 0.0
				ELSE (
					GREATEST(
						COALESCE(a.allocated, 0) - COALESCE(e.used, 0),
						0
					)::FLOAT /
					COALESCE(a.allocated, 0) * 100
				)
			END AS available_percentage
		FROM active_jars j
		LEFT JOIN allocation_totals a
			ON j.id = a.jar_id
		LEFT JOIN expense_totals e
			ON j.id = e.jar_id
		ORDER BY j.created_at, j.id
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jars []*ports.JarSummary

	for rows.Next() {
		var jar ports.JarSummary

		if err := rows.Scan(
			&jar.ID,
			&jar.Name,
			&jar.AllocationType,
			&jar.AllocationValue,
			&jar.Allocated,
			&jar.Used,
			&jar.Available,
			&jar.UsedPercentage,
			&jar.AvailablePercentage,
		); err != nil {
			return nil, err
		}

		jars = append(jars, &jar)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jars, nil
}

// GetGoalSummaries returns all non-archived goals with progress.
func (r *dashboardRepository) GetGoalSummaries(ctx context.Context) ([]*ports.GoalSummary, error) {
	db := dbExecutor(ctx, r.db)

	const query = `
		WITH active_goals AS (
			SELECT
				id,
				name,
				target,
				deadline,
				created_at
			FROM goals
			WHERE is_archived = FALSE
		),
		contribution_totals AS (
			SELECT
				goal_id,
				SUM(amount) AS contributed
			FROM contributions
			GROUP BY goal_id
		)
		SELECT
			g.id,
			g.name,
			g.target,
			COALESCE(c.contributed, 0) AS contributed,
			GREATEST(
				g.target - COALESCE(c.contributed, 0),
				0
			) AS remaining,
			LEAST(
				CASE
					WHEN g.target = 0 THEN 0.0
					ELSE (
						COALESCE(c.contributed, 0)::FLOAT /
						g.target * 100
					)
				END,
				100.0
			) AS progress,
			g.deadline
		FROM active_goals g
		LEFT JOIN contribution_totals c
			ON c.goal_id = g.id
		ORDER BY g.created_at, g.id
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []*ports.GoalSummary

	for rows.Next() {
		var goal ports.GoalSummary

		if err := rows.Scan(
			&goal.ID,
			&goal.Name,
			&goal.Target,
			&goal.Contributed,
			&goal.Remaining,
			&goal.Progress,
			&goal.Deadline,
		); err != nil {
			return nil, err
		}

		goals = append(goals, &goal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}

// GetRecentTransactions returns the N most recent transactions.
func (r *dashboardRepository) GetRecentTransactions(ctx context.Context, limit int) ([]*ports.RecentTransaction, error) {
	db := dbExecutor(ctx, r.db)

	const query = `
		SELECT
			id,
			name,
			type,
			category,
			amount,
			occurred_at
		FROM transactions
		ORDER BY occurred_at DESC, id DESC
		LIMIT $1
	`

	rows, err := db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*ports.RecentTransaction

	for rows.Next() {
		var transaction ports.RecentTransaction

		if err := rows.Scan(
			&transaction.ID,
			&transaction.Name,
			&transaction.Type,
			&transaction.Category,
			&transaction.Amount,
			&transaction.OccurredAt,
		); err != nil {
			return nil, err
		}

		transactions = append(transactions, &transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
