// Package transaction contains the core domain model for an transaction
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
	ID             uuid.UUID
	Name           string
	Type           TransactionType
	Category       TransactionCategory
	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	JarID          uuid.UUID
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
	fromAccountID uuid.UUID,
	toAccountID uuid.UUID,
	jarID uuid.UUID,
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

	// Type-specific validation rules
	switch transactionType {
	case TransactionTypeIncome:
		if toAccountID == uuid.Nil {
			return nil, ErrTargetAccountRequired
		}
	case TransactionTypeExpense:
		if fromAccountID == uuid.Nil {
			return nil, ErrSourceAccountRequired
		}
	case TransactionTypeTransfer:
		if fromAccountID == uuid.Nil || toAccountID == uuid.Nil {
			return nil, ErrInvalidAccount
		}
		if fromAccountID == toAccountID {
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
