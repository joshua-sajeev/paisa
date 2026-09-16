package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
)

// GoalRepository defines the persistence port for goals.
type GoalRepository interface {
	// Create creates a new goal.
	Create(ctx context.Context, g *goal.Goal) error
	
	// List gets all goals.
	List(ctx context.Context) ([]*goal.Goal, error)
	
	// FindByID gets a goal by its ID.
	FindByID(ctx context.Context, id uuid.UUID) (*goal.Goal, error)
	
	// Save persists the current state of an existing goal.
	Save(ctx context.Context, g *goal.Goal) error
}
