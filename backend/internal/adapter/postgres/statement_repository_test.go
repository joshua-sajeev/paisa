package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
	"github.com/joshu-sajeev/paisa/internal/ports"
	"github.com/stretchr/testify/require"
)

func createTestTxn(
	t *testing.T,
	name string,
	typ transaction.TransactionType,
	cat transaction.TransactionCategory,
	from, to, jarID *uuid.UUID,
	amount int64,
	at time.Time,
) {
	t.Helper()

	txn, err := transaction.NewTransaction(name, typ, cat, from, to, jarID, amount, at, false)
	require.NoError(t, err)
	require.NoError(t, transactionRepo.Create(ctx, txn))
}

func TestTransactionRepository_StatementSummaryCountAndRunningBalance(t *testing.T) {
	t.Cleanup(func() { truncateTables(t, ctx, db) })

	acc, err := account.NewAccount("Stmt Main", false, "")
	require.NoError(t, err)
	require.NoError(t, accountRepo.Create(ctx, acc))

	other, err := account.NewAccount("Stmt Other", false, "")
	require.NoError(t, err)
	require.NoError(t, accountRepo.Create(ctx, other))

	j, err := jar.NewJar("Fun", jar.AllocationTypePercentage, 10, "")
	require.NoError(t, err)
	require.NoError(t, jarRepo.Create(ctx, j))

	at := func(m time.Month, d, h int) time.Time {
		return time.Date(2026, m, d, h, 0, 0, 0, time.UTC)
	}

	// History (balance after each, for acc):
	//   Aug 20 income   +1000 -> 1000   (before the range)
	//   Sep 02 expense   -200 ->  800
	//   Sep 09 expense   -300 ->  500   (jar "Fun")
	//   Sep 15 transfer  -100 ->  400   (to other)
	//   Sep 20 transfer  +50  ->  450   (from other)
	//   Sep 25 income    +700 -> 1150
	//   Oct 05 expense   -150 -> 1000   (after the range)
	cat := transaction.TransactionCategoryOther
	exp := transaction.TransactionCategoryEntertainment
	createTestTxn(t, "Opening salary", transaction.TransactionTypeIncome, cat, nil, &acc.ID, nil, 1000, at(8, 20, 6))
	createTestTxn(t, "Groceries", transaction.TransactionTypeExpense, cat, &acc.ID, nil, nil, 200, at(9, 2, 6))
	createTestTxn(t, "Movie", transaction.TransactionTypeExpense, exp, &acc.ID, nil, &j.ID, 300, at(9, 9, 6))
	createTestTxn(t, "To other", transaction.TransactionTypeTransfer, transaction.TransactionCategoryTransfer, &acc.ID, &other.ID, nil, 100, at(9, 15, 6))
	createTestTxn(t, "From other", transaction.TransactionTypeTransfer, transaction.TransactionCategoryTransfer, &other.ID, &acc.ID, nil, 50, at(9, 20, 6))
	createTestTxn(t, "Bonus", transaction.TransactionTypeIncome, cat, nil, &acc.ID, nil, 700, at(9, 25, 6))
	createTestTxn(t, "Late expense", transaction.TransactionTypeExpense, cat, &acc.ID, nil, nil, 150, at(10, 5, 6))

	from := at(9, 1, 0)
	to := at(10, 1, 0) // exclusive

	t.Run("summary over September", func(t *testing.T) {
		s, err := transactionRepo.SummarizeByAccount(ctx, acc.ID, &from, &to)
		require.NoError(t, err)

		require.Equal(t, int64(1000), s.OpeningBalance)
		require.Equal(t, int64(750), s.Inflow)  // 50 + 700
		require.Equal(t, int64(600), s.Outflow) // 200 + 300 + 100
		require.Equal(t, int64(150), s.Net())
		require.Equal(t, int64(1150), s.ClosingBalance)
	})

	t.Run("open-ended range covers all history", func(t *testing.T) {
		s, err := transactionRepo.SummarizeByAccount(ctx, acc.ID, nil, nil)
		require.NoError(t, err)

		require.Equal(t, int64(0), s.OpeningBalance)
		require.Equal(t, int64(1750), s.Inflow)
		require.Equal(t, int64(750), s.Outflow)
		require.Equal(t, int64(1000), s.ClosingBalance)
	})

	t.Run("empty range has opening == closing", func(t *testing.T) {
		f := at(7, 1, 0)
		u := at(7, 31, 0)
		s, err := transactionRepo.SummarizeByAccount(ctx, acc.ID, &f, &u)
		require.NoError(t, err)

		require.Equal(t, int64(0), s.Inflow)
		require.Equal(t, int64(0), s.Outflow)
		require.Equal(t, s.OpeningBalance, s.ClosingBalance)
	})

	t.Run("count honours filters and ignores pagination", func(t *testing.T) {
		n, err := transactionRepo.CountByAccount(ctx, acc.ID, ports.ListParams{
			FromDate: &from, ToDate: &to, Limit: 1,
		})
		require.NoError(t, err)
		require.Equal(t, 5, n)

		expense := transaction.TransactionTypeExpense
		n, err = transactionRepo.CountByAccount(ctx, acc.ID, ports.ListParams{
			FromDate: &from, ToDate: &to, Type: &expense,
		})
		require.NoError(t, err)
		require.Equal(t, 2, n)

		n, err = transactionRepo.CountByAccount(ctx, acc.ID, ports.ListParams{JarID: &j.ID})
		require.NoError(t, err)
		require.Equal(t, 1, n)

		search := "movie"
		n, err = transactionRepo.CountByAccount(ctx, acc.ID, ports.ListParams{Search: &search})
		require.NoError(t, err)
		require.Equal(t, 1, n)
	})

	t.Run("running balance is true balance under filters and pages", func(t *testing.T) {
		items, err := transactionRepo.ListByAccount(ctx, acc.ID, ports.ListParams{
			FromDate: &from, ToDate: &to, Limit: 2, Offset: 0,
		})
		require.NoError(t, err)
		require.Len(t, items, 2)

		// Newest first: Bonus (1150), From other (450).
		require.Equal(t, "Bonus", items[0].Name)
		require.Equal(t, int64(700), items[0].Amount)
		require.Equal(t, int64(1150), *items[0].AccountBalance)
		require.Equal(t, "From other", items[1].Name)
		require.Equal(t, int64(450), *items[1].AccountBalance)

		page2, err := transactionRepo.ListByAccount(ctx, acc.ID, ports.ListParams{
			FromDate: &from, ToDate: &to, Limit: 2, Offset: 2,
		})
		require.NoError(t, err)
		require.Len(t, page2, 2)
		require.Equal(t, "To other", page2[0].Name)
		require.Equal(t, int64(-100), page2[0].Amount)
		require.Equal(t, int64(400), *page2[0].AccountBalance)

		// Filtering to the jar must not change the balance shown for that row.
		jarItems, err := transactionRepo.ListByAccount(ctx, acc.ID, ports.ListParams{
			JarID: &j.ID, Limit: 10,
		})
		require.NoError(t, err)
		require.Len(t, jarItems, 1)
		require.Equal(t, "Movie", jarItems[0].Name)
		require.Equal(t, int64(500), *jarItems[0].AccountBalance)
	})
}
