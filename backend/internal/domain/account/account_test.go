package account_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		{
			name:     "valid name",
			input:    "Checking",
			wantName: "Checking",
		},
		{
			name:     "trims whitespace",
			input:    "  Checking  ",
			wantName: "Checking",
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: account.ErrInvalidName,
		},
		{
			name:    "whitespace only name",
			input:   "   ",
			wantErr: account.ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := account.NewAccount(tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewAccount() error = %v, wantErr = %v", err, tt.wantErr)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Fatalf("NewAccount() returned account, want nil")
				}
				return
			}

			if got == nil {
				t.Fatal("NewAccount() returned nil account")
			}

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.IsArchived {
				t.Error("new account should not be archived")
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
