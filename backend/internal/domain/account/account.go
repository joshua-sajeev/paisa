// Package account contains the core domain model for an account
package account

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID         uuid.UUID
	Name       string
	IsArchived bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewAccount(name string) (*Account, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrInvalidName
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	return &Account{
		ID:         uuid.New(),
		Name:       name,
		IsArchived: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
