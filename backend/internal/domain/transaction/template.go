package transaction

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Template struct {
	ID            uuid.UUID
	Name          string
	Type          TransactionType
	Category      TransactionCategory
	FromAccountID *uuid.UUID
	ToAccountID   *uuid.UUID
	JarID         *uuid.UUID
	Amount        *int64
	IsArchived    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTemplate creates and validates a new Template entity. Category is optional
// only for source compatibility with older callers; new API callers must provide it.
func NewTemplate(name string, transactionType TransactionType, fromAccountID, toAccountID, jarID uuid.UUID, amount *int64, categories ...TransactionCategory) (*Template, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrTemplateInvalidName
	}
	if !transactionType.IsValid() {
		return nil, ErrInvalidTransactionType
	}
	category := TransactionCategoryOther
	if len(categories) > 0 {
		category = categories[0]
	}
	if !category.IsValid() {
		return nil, ErrInvalidCategory
	}
	if amount != nil && *amount <= 0 {
		return nil, ErrTemplateInvalidAmount
	}

	now := time.Now().UTC().Truncate(time.Microsecond)
	return &Template{
		ID: uuid.New(), Name: name, Type: transactionType, Category: category,
		FromAccountID: uuidPtr(fromAccountID), ToAccountID: uuidPtr(toAccountID),
		JarID: uuidPtr(jarID), Amount: amount, IsArchived: false,
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

func uuidPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	value := id
	return &value
}
