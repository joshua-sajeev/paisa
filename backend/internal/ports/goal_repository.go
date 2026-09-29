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

	// Delete removes a goal and its contributions.
	Delete(ctx context.Context, id uuid.UUID) error

	// AddContribution adds a new contribution to a goal.
	AddContribution(ctx context.Context, c *goal.Contribution) error

	// ListContributions lists all contributions for a goal.
	ListContributions(ctx context.Context, goalID uuid.UUID) ([]*goal.Contribution, error)
}
