package goal_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
)

func TestNewGoal(t *testing.T) {
	deadline := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		inputName  string
		target     int64
		deadline   time.Time
		wantName   string
		wantTarget int64
		wantErr    error
	}{
		{
			name:       "valid goal",
			inputName:  "Emergency Fund",
			target:     100000,
			deadline:   deadline,
			wantName:   "Emergency Fund",
			wantTarget: 100000,
		},
		{
			name:       "trims whitespace",
			inputName:  "  Emergency Fund  ",
			target:     100000,
			deadline:   deadline,
			wantName:   "Emergency Fund",
			wantTarget: 100000,
		},
		{
			name:      "empty name",
			inputName: "",
			target:    100000,
			deadline:  deadline,
			wantErr:   goal.ErrInvalidName,
		},
		{
			name:      "whitespace only name",
			inputName: "   ",
			target:    100000,
			deadline:  deadline,
			wantErr:   goal.ErrInvalidName,
		},
		{
			name:      "zero target",
			inputName: "Emergency Fund",
			target:    0,
			deadline:  deadline,
			wantErr:   goal.ErrInvalidTarget,
		},
		{
			name:      "negative target",
			inputName: "Emergency Fund",
			target:    -100,
			deadline:  deadline,
			wantErr:   goal.ErrInvalidTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := goal.NewGoal(
				tt.inputName,
				tt.target,
				tt.deadline,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"NewGoal() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Fatal("NewGoal() returned goal, want nil")
				}
				return
			}

			if got == nil {
				t.Fatal("NewGoal() returned nil goal")
			}

			if got.Name != tt.wantName {
				t.Errorf(
					"Name = %q, want %q",
					got.Name,
					tt.wantName,
				)
			}

			if got.Target != tt.wantTarget {
				t.Errorf(
					"Target = %d, want %d",
					got.Target,
					tt.wantTarget,
				)
			}

			if !got.Deadline.Equal(tt.deadline) {
				t.Errorf(
					"Deadline = %v, want %v",
					got.Deadline,
					tt.deadline,
				)
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.IsArchived {
				t.Error("new goal should not be archived")
			}

			if got.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}

			if got.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}

			if !got.CreatedAt.Equal(got.UpdatedAt) {
				t.Errorf(
					"CreatedAt = %v, UpdatedAt = %v, want equal timestamps",
					got.CreatedAt,
					got.UpdatedAt,
				)
			}
		})
	}
}

func TestNewContribution(t *testing.T) {
	goalID := uuid.New()
	occurredAt := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		goalID     uuid.UUID
		amount     int64
		occurredAt time.Time
		wantErr    error
	}{
		{
			name:       "valid contribution",
			goalID:     goalID,
			amount:     5000,
			occurredAt: occurredAt,
		},
		{
			name:       "zero amount",
			goalID:     goalID,
			amount:     0,
			occurredAt: occurredAt,
			wantErr:    goal.ErrInvalidAmount,
		},
		{
			name:       "negative amount",
			goalID:     goalID,
			amount:     -100,
			occurredAt: occurredAt,
			wantErr:    goal.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := goal.NewContribution(tt.goalID, tt.amount, tt.occurredAt)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewContribution() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				return
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.GoalID != tt.goalID {
				t.Errorf("GoalID = %v, want %v", got.GoalID, tt.goalID)
			}

			if got.Amount != tt.amount {
				t.Errorf("Amount = %d, want %d", got.Amount, tt.amount)
			}

			if !got.OccurredAt.Equal(tt.occurredAt) {
				t.Errorf("OccurredAt = %v, want %v", got.OccurredAt, tt.occurredAt)
			}

			if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
				t.Error("timestamps should be set")
			}
		})
	}
}
