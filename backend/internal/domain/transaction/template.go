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
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	JarID         uuid.UUID
	Amount        *int64
	IsArchived    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTemplate creates and validates a new Template entity.
func NewTemplate(
	name string,
	transactionType TransactionType,
	fromAccountID uuid.UUID,
	toAccountID uuid.UUID,
	jarID uuid.UUID,
	amount *int64,
) (*Template, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrTemplateInvalidName
	}

	if !transactionType.IsValid() {
		return nil, ErrInvalidTransactionType
	}

	if amount != nil && *amount <= 0 {
		return nil, ErrTemplateInvalidAmount
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	return &Template{
		ID:            uuid.New(),
		Name:          name,
		Type:          transactionType,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		JarID:         jarID,
		Amount:        amount,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}
