package jar_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

func TestNewJar(t *testing.T) {
	tests := []struct {
		name            string
		inputName       string
		allocationType  jar.AllocationType
		allocationValue int64
		wantName        string
		wantErr         error
	}{
		{
			name:            "valid percentage jar",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantName:        "Needs",
		},
		{
			name:            "valid minimum percentage",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 1,
			wantName:        "Needs",
		},
		{
			name:            "valid maximum percentage",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 100,
			wantName:        "Needs",
		},
		{
			name:            "valid remainder jar",
			inputName:       "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 0,
			wantName:        "Remainder",
		},
		{
			name:            "trims whitespace",
			inputName:       "  Needs  ",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantName:        "Needs",
		},
		{
			name:            "empty name",
			inputName:       "",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantErr:         jar.ErrInvalidName,
		},
		{
			name:            "whitespace only name",
			inputName:       "   ",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 50,
			wantErr:         jar.ErrInvalidName,
		},
		{
			name:            "invalid allocation type",
			inputName:       "Needs",
			allocationType:  "invalid",
			allocationValue: 50,
			wantErr:         jar.ErrInvalidAllocationType,
		},
		{
			name:            "percentage below minimum",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 0,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "percentage above maximum",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: 101,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "negative percentage",
			inputName:       "Needs",
			allocationType:  jar.AllocationTypePercentage,
			allocationValue: -1,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "remainder must be zero",
			inputName:       "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: 1,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
		{
			name:            "remainder cannot be negative",
			inputName:       "Remainder",
			allocationType:  jar.AllocationTypeRemainder,
			allocationValue: -1,
			wantErr:         jar.ErrInvalidAllocationVal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jar.NewJar(
				tt.inputName,
				tt.allocationType,
				tt.allocationValue,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"NewJar() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Fatal("NewJar() returned jar, want nil")
				}
				return
			}

			if got == nil {
				t.Fatal("NewJar() returned nil jar")
			}

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}

			if got.AllocationType != tt.allocationType {
				t.Errorf(
					"AllocationType = %q, want %q",
					got.AllocationType,
					tt.allocationType,
				)
			}

			if got.AllocationValue != tt.allocationValue {
				t.Errorf(
					"AllocationValue = %d, want %d",
					got.AllocationValue,
					tt.allocationValue,
				)
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.IsArchived {
				t.Error("new jar should not be archived")
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
