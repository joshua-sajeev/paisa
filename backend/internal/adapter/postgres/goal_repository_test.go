package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
)

func TestGoalRepository_Contributions(t *testing.T) {
	ctx := context.Background()
	truncateTables(t, ctx, db)

	// Create a goal first
	g, err := goal.NewGoal("Vacation", 100000, time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("new goal: %v", err)
	}
	if err := goalRepo.Create(ctx, g); err != nil {
		t.Fatalf("create goal: %v", err)
	}

	t.Run("Add and List Contributions", func(t *testing.T) {
		c1, _ := goal.NewContribution(g.ID, 5000, time.Now().Add(-1*time.Hour))
		c2, _ := goal.NewContribution(g.ID, 3000, time.Now())

		if err := goalRepo.AddContribution(ctx, c1); err != nil {
			t.Fatalf("add contribution 1: %v", err)
		}
		if err := goalRepo.AddContribution(ctx, c2); err != nil {
			t.Fatalf("add contribution 2: %v", err)
		}

		contributions, err := goalRepo.ListContributions(ctx, g.ID)
		if err != nil {
			t.Fatalf("list contributions: %v", err)
		}

		if len(contributions) != 2 {
			t.Errorf("len(contributions) = %d, want 2", len(contributions))
		}

		// Ordered by occurred_at DESC
		if contributions[0].ID != c2.ID {
			t.Errorf("contributions[0].ID = %v, want %v", contributions[0].ID, c2.ID)
		}
	})

	t.Run("List for non-existent goal returns empty", func(t *testing.T) {
		contributions, err := goalRepo.ListContributions(ctx, uuid.New())
		if err != nil {
			t.Fatalf("list contributions: %v", err)
		}

		if len(contributions) != 0 {
			t.Errorf("len(contributions) = %d, want 0", len(contributions))
		}
	})
}

func TestGoalRepository_Delete(t *testing.T) {
	ctx := context.Background()
	truncateTables(t, ctx, db)

	t.Run("Delete goal with 0 contributions", func(t *testing.T) {
		g, err := goal.NewGoal(
			"Emergency",
			50000,
			time.Now().Add(24*time.Hour),
		)
		if err != nil {
			t.Fatalf("new goal: %v", err)
		}

		if err := goalRepo.Create(ctx, g); err != nil {
			t.Fatalf("create goal: %v", err)
		}

		if err := goalRepo.Delete(ctx, g.ID); err != nil {
			t.Fatalf("delete goal: %v", err)
		}

		_, err = goalRepo.FindByID(ctx, g.ID)
		if !errors.Is(err, goal.ErrGoalNotFound) {
			t.Errorf("expected ErrGoalNotFound, got %v", err)
		}
	})

	t.Run("Delete goal with contributions (cascades)", func(t *testing.T) {
		g, err := goal.NewGoal(
			"Retirement",
			500000,
			time.Now().Add(24*time.Hour),
		)
		if err != nil {
			t.Fatalf("new goal: %v", err)
		}

		if err := goalRepo.Create(ctx, g); err != nil {
			t.Fatalf("create goal: %v", err)
		}

		c, _ := goal.NewContribution(
			g.ID,
			500000,
			time.Now(),
		)

		if err := goalRepo.AddContribution(ctx, c); err != nil {
			t.Fatalf("add contribution: %v", err)
		}

		if err := goalRepo.Delete(ctx, g.ID); err != nil {
			t.Fatalf("delete goal: %v", err)
		}

		_, err = goalRepo.FindByID(ctx, g.ID)
		if !errors.Is(err, goal.ErrGoalNotFound) {
			t.Errorf("expected ErrGoalNotFound, got %v", err)
		}

		contributions, err := goalRepo.ListContributions(ctx, g.ID)
		if err != nil {
			t.Fatalf("list contributions: %v", err)
		}

		if len(contributions) != 0 {
			t.Errorf(
				"expected 0 contributions after cascade delete, got %d",
				len(contributions),
			)
		}
	})

	t.Run("Delete non-existent goal returns ErrGoalNotFound", func(t *testing.T) {
		err := goalRepo.Delete(ctx, uuid.New())

		if !errors.Is(err, goal.ErrGoalNotFound) {
			t.Errorf("expected ErrGoalNotFound, got %v", err)
		}
	})
}
