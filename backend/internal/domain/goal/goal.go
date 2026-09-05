// Package goal contains the core domain model for a goal
package goal

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Goal struct {
	ID         uuid.UUID
	Name       string
	Target     int64
	IsArchived bool
	Deadline   time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewGoal(name string, target int64) (*Goal, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, ErrInvalidName
	}

	if target <= 0 {
		return nil, ErrInvalidTarget
	}

	now := time.Now().UTC().Truncate(time.Microsecond)

	return &Goal{
		ID:         uuid.New(),
		Name:       name,
		Target:     target,
		IsArchived: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}
