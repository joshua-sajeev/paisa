package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
	"github.com/joshu-sajeev/paisa/internal/ports"
)

// GoalService handles goal application use cases.
type GoalService struct {
	repo   ports.GoalRepository
	logger *slog.Logger
}

// NewGoalService creates a new GoalService.
func NewGoalService(repo ports.GoalRepository, logger *slog.Logger) *GoalService {
	return &GoalService{
		repo:   repo,
		logger: logger,
	}
}

// Create creates a new goal.
func (s *GoalService) Create(ctx context.Context, name string, target int64, deadline time.Time) (*goal.Goal, error) {
	s.logger.DebugContext(ctx, "creating new goal", slog.String("name", name))

	newGoal, err := goal.NewGoal(name, target, deadline)
	if err != nil {
		s.logger.WarnContext(ctx, "invalid goal", slog.String("error", err.Error()))
		return nil, err
	}

	if err := s.repo.Create(ctx, newGoal); err != nil {
		s.logger.ErrorContext(ctx, "failed to create goal", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "goal created successfully", slog.String("id", newGoal.ID.String()))
	return newGoal, nil
}

// List gets all goals.
func (s *GoalService) List(ctx context.Context) ([]*goal.Goal, error) {
	s.logger.DebugContext(ctx, "listing goals")

	goals, err := s.repo.List(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "repository list failed", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "goals listed", slog.Int("count", len(goals)))
	return goals, nil
}

// Update updates an existing goal.
func (s *GoalService) Update(ctx context.Context, id uuid.UUID, name *string, target *int64, deadline *time.Time, isArchived *bool) error {
	s.logger.DebugContext(ctx, "updating goal", slog.String("id", id.String()))

	g, err := s.repo.FindByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to find goal for update", slog.String("error", err.Error()))
		return err
	}

	changed := false

	if name != nil && *name != g.Name {
		if err := g.Rename(*name); err != nil {
			return err
		}
		changed = true
	}

	if target != nil && *target != g.Target {
		if err := g.UpdateTarget(*target); err != nil {
			return err
		}
		changed = true
	}

	if deadline != nil && !deadline.Equal(g.Deadline) {
		g.UpdateDeadline(*deadline)
		changed = true
	}

	if isArchived != nil && *isArchived != g.IsArchived {
		if *isArchived {
			g.Archive()
		} else {
			g.Unarchive()
		}
		changed = true
	}

	if !changed {
		return nil
	}

	if err := s.repo.Save(ctx, g); err != nil {
		s.logger.ErrorContext(ctx, "repository goal save failed", slog.String("error", err.Error()))
		return err
	}

	s.logger.InfoContext(ctx, "goal updated successfully", slog.String("id", id.String()))
	return nil
}
