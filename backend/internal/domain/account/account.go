// Package account contains the core domain model for an account.
package account

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID         uuid.UUID
	Name       string
	Balance    int64
	IsArchived bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewAccount creates a new active account.
func NewAccount(name string) (*Account, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrInvalidName
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	return &Account{
		ID:         uuid.New(),
		Name:       name,
		Balance:    0,
		IsArchived: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Rename changes the account name.
func (a *Account) Rename(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrInvalidName
	}

	if a.Name == name {
		return nil
	}

	a.Name = name
	a.touch()

	return nil
}

// Archive archives the account.
func (a *Account) Archive() error {
	if a.IsArchived {
		return ErrAccountAlreadyArchived
	}

	a.IsArchived = true
	a.touch()

	return nil
}

// Unarchive restores the account to an active state.
func (a *Account) Unarchive() error {
	if !a.IsArchived {
		return ErrAccountNotArchived
	}

	a.IsArchived = false
	a.touch()

	return nil
}

func (a *Account) touch() {
	a.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
}

// UpdateBalance changes the account balance by the given amount.
func (a *Account) UpdateBalance(amount int64) {
	a.Balance += amount
	a.touch()
}
