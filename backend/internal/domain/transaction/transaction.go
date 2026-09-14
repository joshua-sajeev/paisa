// Package transaction contains the core domain model for a transaction.
package transaction

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

func (tt TransactionType) IsValid() bool {
	switch tt {
	case TransactionTypeIncome,
		TransactionTypeExpense,
		TransactionTypeTransfer:
		return true
	default:
		return false
	}
}

type Transaction struct {
	ID            uuid.UUID
	Name          string
	Type          TransactionType
	Category      TransactionCategory
	FromAccountID *uuid.UUID
	ToAccountID   *uuid.UUID
	JarID         *uuid.UUID

	// Amount is stored in the smallest currency unit (paise for INR).
	// Minimum value: 1 paisa (₹0.01).
	Amount         int64
	OccurredAt     time.Time
	IsMasterIncome bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewTransaction(
	name string,
	transactionType TransactionType,
	category TransactionCategory,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	jarID *uuid.UUID,
	amount int64,
	occurredAt time.Time,
	isMasterIncome bool,
) (*Transaction, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidName
	}

	if !transactionType.IsValid() {
		return nil, ErrInvalidTransactionType
	}

	if !category.IsValid() {
		return nil, ErrInvalidCategory
	}

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Type-specific validation rules.
	switch transactionType {
	case TransactionTypeIncome:
		if toAccountID == nil {
			return nil, ErrTargetAccountRequired
		}

	case TransactionTypeExpense:
		if fromAccountID == nil {
			return nil, ErrSourceAccountRequired
		}

	case TransactionTypeTransfer:
		if fromAccountID == nil || toAccountID == nil {
			return nil, ErrInvalidAccount
		}

		if *fromAccountID == *toAccountID {
			return nil, ErrInvalidTransfer
		}
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	if occurredAt.IsZero() {
		occurredAt = now
	} else {
		occurredAt = occurredAt.UTC().Truncate(time.Microsecond)
	}

	return &Transaction{
		ID:             uuid.New(),
		Name:           name,
		Type:           transactionType,
		Category:       category,
		FromAccountID:  fromAccountID,
		ToAccountID:    toAccountID,
		JarID:          jarID,
		Amount:         amount,
		OccurredAt:     occurredAt,
		IsMasterIncome: isMasterIncome,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (t *Transaction) Update(
	name string,
	category TransactionCategory,
	fromAccountID *uuid.UUID,
	toAccountID *uuid.UUID,
	jarID *uuid.UUID,
	amount int64,
	occurredAt time.Time,
	isMasterIncome bool,
) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidName
	}

	if !category.IsValid() {
		return ErrInvalidCategory
	}

	if amount <= 0 {
		return ErrInvalidAmount
	}

	switch t.Type {
	case TransactionTypeIncome:
		if toAccountID == nil {
			return ErrTargetAccountRequired
		}

	case TransactionTypeExpense:
		if fromAccountID == nil {
			return ErrSourceAccountRequired
		}

	case TransactionTypeTransfer:
		if fromAccountID == nil || toAccountID == nil {
			return ErrInvalidAccount
		}

		if *fromAccountID == *toAccountID {
			return ErrInvalidTransfer
		}
	}

	t.Name = name
	t.Category = category
	t.FromAccountID = fromAccountID
	t.ToAccountID = toAccountID
	t.JarID = jarID
	t.Amount = amount

	if !occurredAt.IsZero() {
		t.OccurredAt = occurredAt.UTC().Truncate(time.Microsecond)
	}

	t.IsMasterIncome = isMasterIncome
	t.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	return nil
}

// HasAllocationChanged reports whether amount, master income status, or jar assignment changed,
// indicating allocations need to be recalculated.
func (t *Transaction) HasAllocationChanged(
	oldAmount int64,
	oldIsMasterIncome bool,
	oldJarID *uuid.UUID,
) bool {
	if t.Amount != oldAmount || t.IsMasterIncome != oldIsMasterIncome {
		return true
	}

	if (t.JarID == nil) != (oldJarID == nil) {
		return true
	}

	if t.JarID != nil && oldJarID != nil && *t.JarID != *oldJarID {
		return true
	}

	return false
}
