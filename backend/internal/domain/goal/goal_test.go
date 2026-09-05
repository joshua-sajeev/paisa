package goal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/goal"
)

func TestNewGoal(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		target    int64
		wantName  string
		wantErr   error
	}{
		{
			name:      "valid goal",
			inputName: "Emergency Fund",
			target:    100000,
			wantName:  "Emergency Fund",
		},
		{
			name:      "trims whitespace",
			inputName: "  Emergency Fund  ",
			target:    100000,
			wantName:  "Emergency Fund",
		},
		{
			name:      "empty name",
			inputName: "",
			target:    100000,
			wantErr:   goal.ErrInvalidName,
		},
		{
			name:      "whitespace only name",
			inputName: "   ",
			target:    100000,
			wantErr:   goal.ErrInvalidName,
		},
		{
			name:      "zero target",
			inputName: "Emergency Fund",
			target:    0,
			wantErr:   goal.ErrInvalidTarget,
		},
		{
			name:      "negative target",
			inputName: "Emergency Fund",
			target:    -100,
			wantErr:   goal.ErrInvalidTarget,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := goal.NewGoal(tt.inputName, tt.target)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewGoal() error = %v, wantErr = %v", err, tt.wantErr)
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
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}

			if got.Target != tt.target {
				t.Errorf("Target = %d, want %d", got.Target, tt.target)
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
		})
	}
}
