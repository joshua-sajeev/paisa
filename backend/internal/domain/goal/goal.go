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

func NewGoal(name string, target int64, deadline time.Time) (*Goal, error) {
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
		Deadline:   deadline,
		IsArchived: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (g *Goal) Rename(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return ErrInvalidName
	}

	if g.Name == name {
		return nil
	}

	g.Name = name
	g.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	return nil
}

func (g *Goal) UpdateTarget(target int64) error {
	if target <= 0 {
		return ErrInvalidTarget
	}
	g.Target = target
	g.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	return nil
}

func (g *Goal) UpdateDeadline(deadline time.Time) {
	g.Deadline = deadline
	g.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
}

func (g *Goal) Archive() {
	g.IsArchived = true
	g.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
}

func (g *Goal) Unarchive() {
	g.IsArchived = false
	g.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
}
