package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
	"log/slog"
	"os"
)

type mockGoalRepository struct {
	goals         map[uuid.UUID]*goal.Goal
	contributions map[uuid.UUID][]*goal.Contribution
}

func newMockGoalRepository() *mockGoalRepository {
	return &mockGoalRepository{
		goals:         make(map[uuid.UUID]*goal.Goal),
		contributions: make(map[uuid.UUID][]*goal.Contribution),
	}
}

func (m *mockGoalRepository) Create(ctx context.Context, g *goal.Goal) error {
	m.goals[g.ID] = g
	return nil
}

func (m *mockGoalRepository) List(ctx context.Context) ([]*goal.Goal, error) {
	goals := make([]*goal.Goal, 0, len(m.goals))
	for _, g := range m.goals {
		goals = append(goals, g)
	}
	return goals, nil
}

func (m *mockGoalRepository) FindByID(ctx context.Context, id uuid.UUID) (*goal.Goal, error) {
	g, ok := m.goals[id]
	if !ok {
		return nil, goal.ErrGoalNotFound
	}
	return g, nil
}

func (m *mockGoalRepository) Save(ctx context.Context, g *goal.Goal) error {
	m.goals[g.ID] = g
	return nil
}

func (m *mockGoalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.goals[id]; !ok {
		return goal.ErrGoalNotFound
	}
	delete(m.goals, id)
	delete(m.contributions, id)
	return nil
}

func (m *mockGoalRepository) AddContribution(ctx context.Context, c *goal.Contribution) error {
	m.contributions[c.GoalID] = append(m.contributions[c.GoalID], c)
	return nil
}

func (m *mockGoalRepository) ListContributions(ctx context.Context, goalID uuid.UUID) ([]*goal.Contribution, error) {
	return m.contributions[goalID], nil
}

func TestGoalService_AddContribution(t *testing.T) {
	repo := newMockGoalRepository()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := application.NewGoalService(repo, logger)

	ctx := context.Background()
	g, _ := goal.NewGoal("Car", 500000, time.Now())
	repo.Create(ctx, g)

	t.Run("success", func(t *testing.T) {
		amount := int64(10000)
		occurredAt := time.Now()
		c, err := service.AddContribution(ctx, g.ID, amount, occurredAt)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if c.Amount != amount {
			t.Errorf("amount = %d, want %d", c.Amount, amount)
		}
	})

	t.Run("goal not found", func(t *testing.T) {
		_, err := service.AddContribution(ctx, uuid.New(), 10000, time.Now())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGoalService_Delete(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	t.Run("success when 0 contributions", func(t *testing.T) {
		repo := newMockGoalRepository()
		service := application.NewGoalService(repo, logger)
		g, _ := goal.NewGoal("Laptop", 100000, time.Now())
		_ = repo.Create(ctx, g)

		err := service.Delete(ctx, g.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if _, err := repo.FindByID(ctx, g.ID); err != goal.ErrGoalNotFound {
			t.Errorf("expected goal to be deleted from repo")
		}
	})

	t.Run("fails when in-progress contributions exist", func(t *testing.T) {
		repo := newMockGoalRepository()
		service := application.NewGoalService(repo, logger)
		g, _ := goal.NewGoal("Laptop", 100000, time.Now())
		_ = repo.Create(ctx, g)
		c, _ := goal.NewContribution(g.ID, 40000, time.Now())
		_ = repo.AddContribution(ctx, c)

		err := service.Delete(ctx, g.ID)
		if err != goal.ErrGoalCannotBeDeleted {
			t.Fatalf("expected ErrGoalCannotBeDeleted, got: %v", err)
		}
		if _, err := repo.FindByID(ctx, g.ID); err != nil {
			t.Errorf("goal should not have been deleted")
		}
	})

	t.Run("success when target is achieved", func(t *testing.T) {
		repo := newMockGoalRepository()
		service := application.NewGoalService(repo, logger)
		g, _ := goal.NewGoal("Laptop", 100000, time.Now())
		_ = repo.Create(ctx, g)
		c1, _ := goal.NewContribution(g.ID, 60000, time.Now())
		c2, _ := goal.NewContribution(g.ID, 50000, time.Now())
		_ = repo.AddContribution(ctx, c1)
		_ = repo.AddContribution(ctx, c2)

		err := service.Delete(ctx, g.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if _, err := repo.FindByID(ctx, g.ID); err != goal.ErrGoalNotFound {
			t.Errorf("expected goal to be deleted from repo")
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := newMockGoalRepository()
		service := application.NewGoalService(repo, logger)

		err := service.Delete(ctx, uuid.New())
		if err != goal.ErrGoalNotFound {
			t.Fatalf("expected ErrGoalNotFound, got: %v", err)
		}
	})
}
