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

func TestAllocationRepository_ListAndSummary(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	// Create test account
	acc, err := account.NewAccount("Main Checking", false, "")
	require.NoError(t, err)
	require.NoError(t, accountRepo.Create(ctx, acc))

	// Create jars
	needsJar, err := jar.NewJar("Needs", jar.AllocationTypePercentage, 50)
	require.NoError(t, err)
	require.NoError(t, jarRepo.Create(ctx, needsJar))

	savingsJar, err := jar.NewJar("Savings", jar.AllocationTypeRemainder, 0)
	require.NoError(t, err)
	require.NoError(t, jarRepo.Create(ctx, savingsJar))

	octDate := time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	novDate := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)

	// Create transactions
	txn1, err := transaction.NewTransaction(
		"October Paycheck",
		transaction.TransactionTypeIncome,
		transaction.TransactionCategoryOther,
		nil,
		&acc.ID,
		nil,
		500000,
		octDate,
		true,
	)
	require.NoError(t, err)
	require.NoError(t, transactionRepo.Create(ctx, txn1))

	txn2, err := transaction.NewTransaction(
		"November Paycheck",
		transaction.TransactionTypeIncome,
		transaction.TransactionCategoryOther,
		nil,
		&acc.ID,
		nil,
		600000,
		novDate,
		true,
	)
	require.NoError(t, err)
	require.NoError(t, transactionRepo.Create(ctx, txn2))

	// Create allocations for txn1
	alloc1 := &transaction.Allocation{
		ID:            uuid.New(),
		TransactionID: txn1.ID,
		JarID:         needsJar.ID,
		Amount:        250000,
		CreatedAt:     octDate,
	}
	require.NoError(t, allocationRepo.Create(ctx, alloc1))

	alloc2 := &transaction.Allocation{
		ID:            uuid.New(),
		TransactionID: txn1.ID,
		JarID:         savingsJar.ID,
		Amount:        250000,
		CreatedAt:     octDate,
	}
	require.NoError(t, allocationRepo.Create(ctx, alloc2))

	// Create allocation for txn2
	alloc3 := &transaction.Allocation{
		ID:            uuid.New(),
		TransactionID: txn2.ID,
		JarID:         needsJar.ID,
		Amount:        300000,
		CreatedAt:     novDate,
	}
	require.NoError(t, allocationRepo.Create(ctx, alloc3))

	// 1. Test List all
	items, total, err := allocationRepo.List(ctx, ports.AllocationListParams{
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, items, 3)

	// Ordered by occurred_at DESC (November first)
	require.Equal(t, txn2.ID, items[0].TransactionID)
	require.Equal(t, "November Paycheck", items[0].TransactionName)
	require.Equal(t, "Needs", items[0].JarName)
	require.Equal(t, jar.AllocationTypePercentage, items[0].AllocationType)
	require.Equal(t, int64(300000), items[0].Amount)

	// 2. Test Search filter
	searchStr := "October"
	searchItems, searchTotal, err := allocationRepo.List(ctx, ports.AllocationListParams{
		Search: &searchStr,
	})
	require.NoError(t, err)
	require.Equal(t, 2, searchTotal)
	require.Len(t, searchItems, 2)

	// 3. Test JarID filter
	jarItems, jarTotal, err := allocationRepo.List(ctx, ports.AllocationListParams{
		JarID: &savingsJar.ID,
	})
	require.NoError(t, err)
	require.Equal(t, 1, jarTotal)
	require.Len(t, jarItems, 1)
	require.Equal(t, savingsJar.ID, jarItems[0].JarID)
	require.Equal(t, "Savings", jarItems[0].JarName)

	// 4. Test AllocationType filter
	allocType := jar.AllocationTypePercentage
	typeItems, typeTotal, err := allocationRepo.List(ctx, ports.AllocationListParams{
		AllocationType: &allocType,
	})
	require.NoError(t, err)
	require.Equal(t, 2, typeTotal)
	require.Len(t, typeItems, 2)

	// 5. Test Date Range filter
	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	dateItems, dateTotal, err := allocationRepo.List(ctx, ports.AllocationListParams{
		FromDate: &from,
		ToDate:   &to,
	})
	require.NoError(t, err)
	require.Equal(t, 1, dateTotal)
	require.Len(t, dateItems, 1)
	require.Equal(t, txn2.ID, dateItems[0].TransactionID)

	// 6. Test Amount filter
	minAmt := int64(260000)
	amtItems, amtTotal, err := allocationRepo.List(ctx, ports.AllocationListParams{
		MinAmount: &minAmt,
	})
	require.NoError(t, err)
	require.Equal(t, 1, amtTotal)
	require.Len(t, amtItems, 1)
	require.Equal(t, int64(300000), amtItems[0].Amount)

	// 7. Test Pagination
	pageItems, pageTotal, err := allocationRepo.List(ctx, ports.AllocationListParams{
		Limit:  1,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Equal(t, 3, pageTotal)
	require.Len(t, pageItems, 1)

	// 8. Test Monthly Summary
	octStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	octEnd := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	// Overall October total (alloc1 + alloc2 = 250k + 250k = 500k)
	octTotal, err := allocationRepo.GetMonthlySummary(ctx, octStart, octEnd, nil)
	require.NoError(t, err)
	require.Equal(t, int64(500000), octTotal)

	// Specific jar in October (Needs = 250k)
	octNeedsTotal, err := allocationRepo.GetMonthlySummary(ctx, octStart, octEnd, &needsJar.ID)
	require.NoError(t, err)
	require.Equal(t, int64(250000), octNeedsTotal)

	// Specific jar in November (Needs = 300k)
	novStart := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	novEnd := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	novNeedsTotal, err := allocationRepo.GetMonthlySummary(ctx, novStart, novEnd, &needsJar.ID)
	require.NoError(t, err)
	require.Equal(t, int64(300000), novNeedsTotal)
}
