package account_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantIconKey string
		wantErr     error
	}{
		{
			name:        "valid name",
			input:       "Checking",
			wantName:    "Checking",
			wantIconKey: "bank",
		},
		{
			name:        "trims whitespace",
			input:       "  Checking  ",
			wantName:    "Checking",
			wantIconKey: "bank",
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
				t.Fatalf(
					"NewAccount() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
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

			if got.IconKey != tt.wantIconKey {
				t.Errorf(
					"IconKey = %q, want %q",
					got.IconKey,
					tt.wantIconKey,
				)
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.Balance != 0 {
				t.Errorf("Balance = %d, want 0", got.Balance)
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

func TestAccountUpdateBalance(t *testing.T) {
	acc, err := account.NewAccount("Checking")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	originalUpdatedAt := acc.UpdatedAt

	time.Sleep(time.Millisecond)

	acc.UpdateBalance(5000)

	if acc.Balance != 5000 {
		t.Errorf("Balance = %d, want %d", acc.Balance, 5000)
	}

	if !acc.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf(
			"UpdatedAt = %v, want after %v",
			acc.UpdatedAt,
			originalUpdatedAt,
		)
	}

	acc.UpdateBalance(-2000)

	if acc.Balance != 3000 {
		t.Errorf("Balance = %d, want %d", acc.Balance, 3000)
	}
}

func TestAccountRename(t *testing.T) {
	tests := []struct {
		name     string
		newName  string
		wantName string
		wantErr  error
	}{
		{
			name:     "valid name",
			newName:  "Savings",
			wantName: "Savings",
		},
		{
			name:     "trims whitespace",
			newName:  "  Savings  ",
			wantName: "Savings",
		},
		{
			name:    "empty name",
			newName: "",
			wantErr: account.ErrInvalidName,
		},
		{
			name:    "whitespace only name",
			newName: "   ",
			wantErr: account.ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := account.NewAccount("Checking")
			if err != nil {
				t.Fatalf("NewAccount() error = %v", err)
			}

			originalUpdatedAt := acc.UpdatedAt

			time.Sleep(time.Millisecond)

			err = acc.Rename(tt.newName)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"Rename() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if acc.Name != "Checking" {
					t.Errorf(
						"Name = %q after failed Rename(), want %q",
						acc.Name,
						"Checking",
					)
				}

				if !acc.UpdatedAt.Equal(originalUpdatedAt) {
					t.Errorf(
						"UpdatedAt changed after failed Rename(): got=%v, want=%v",
						acc.UpdatedAt,
						originalUpdatedAt,
					)
				}

				return
			}

			if acc.Name != tt.wantName {
				t.Errorf(
					"Name = %q, want %q",
					acc.Name,
					tt.wantName,
				)
			}

			if !acc.UpdatedAt.After(originalUpdatedAt) {
				t.Errorf(
					"UpdatedAt = %v, want after %v",
					acc.UpdatedAt,
					originalUpdatedAt,
				)
			}
		})
	}
}

func TestAccountArchive(t *testing.T) {
	acc, err := account.NewAccount("Checking")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	originalUpdatedAt := acc.UpdatedAt

	time.Sleep(time.Millisecond)

	if err := acc.Archive(); err != nil {
		t.Fatalf("Archive() error = %v", err)
	}

	if !acc.IsArchived {
		t.Error("IsArchived = false, want true")
	}

	if !acc.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf(
			"UpdatedAt = %v, want after %v",
			acc.UpdatedAt,
			originalUpdatedAt,
		)
	}
}

func TestAccountArchive_AlreadyArchived(t *testing.T) {
	acc, err := account.NewAccount("Checking")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	acc.IsArchived = true
	originalUpdatedAt := acc.UpdatedAt

	err = acc.Archive()

	if !errors.Is(err, account.ErrAccountAlreadyArchived) {
		t.Errorf(
			"Archive() error = %v, want %v",
			err,
			account.ErrAccountAlreadyArchived,
		)
	}

	if !acc.IsArchived {
		t.Error("IsArchived = false, want true")
	}

	if !acc.UpdatedAt.Equal(originalUpdatedAt) {
		t.Errorf(
			"UpdatedAt changed after failed Archive(): got=%v, want=%v",
			acc.UpdatedAt,
			originalUpdatedAt,
		)
	}
}

func TestAccountUnarchive(t *testing.T) {
	acc, err := account.NewAccount("Checking")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	acc.IsArchived = true
	originalUpdatedAt := acc.UpdatedAt

	time.Sleep(time.Millisecond)

	if err := acc.Unarchive(); err != nil {
		t.Fatalf("Unarchive() error = %v", err)
	}

	if acc.IsArchived {
		t.Error("IsArchived = true, want false")
	}

	if !acc.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf(
			"UpdatedAt = %v, want after %v",
			acc.UpdatedAt,
			originalUpdatedAt,
		)
	}
}

func TestAccountUnarchive_NotArchived(t *testing.T) {
	acc, err := account.NewAccount("Checking")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	originalUpdatedAt := acc.UpdatedAt

	err = acc.Unarchive()

	if !errors.Is(err, account.ErrAccountNotArchived) {
		t.Errorf(
			"Unarchive() error = %v, want %v",
			err,
			account.ErrAccountNotArchived,
		)
	}

	if acc.IsArchived {
		t.Error("IsArchived = true, want false")
	}

	if !acc.UpdatedAt.Equal(originalUpdatedAt) {
		t.Errorf(
			"UpdatedAt changed after failed Unarchive(): got=%v, want=%v",
			acc.UpdatedAt,
			originalUpdatedAt,
		)
	}
}

func TestAccount_UpdateIcon(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		input    string
		wantIcon string
	}{
		{
			name:     "updates icon",
			initial:  "bank",
			input:    "hdfc",
			wantIcon: "hdfc",
		},
		{
			name:     "trims whitespace",
			initial:  "bank",
			input:    "  hdfc  ",
			wantIcon: "hdfc",
		},
		{
			name:     "empty icon defaults to bank",
			initial:  "hdfc",
			input:    "",
			wantIcon: "bank",
		},
		{
			name:     "whitespace only icon defaults to bank",
			initial:  "hdfc",
			input:    "   ",
			wantIcon: "bank",
		},
		{
			name:     "same icon does not change",
			initial:  "hdfc",
			input:    "hdfc",
			wantIcon: "hdfc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := account.NewAccount("Checking")
			if err != nil {
				t.Fatalf("NewAccount() error = %v", err)
			}

			got.UpdateIcon(tt.initial)

			before := got.UpdatedAt

			got.UpdateIcon(tt.input)

			if got.IconKey != tt.wantIcon {
				t.Errorf(
					"IconKey = %q, want %q",
					got.IconKey,
					tt.wantIcon,
				)
			}

			if tt.input == tt.initial {
				if !got.UpdatedAt.Equal(before) {
					t.Error("UpdatedAt changed when icon was unchanged")
				}
			} else {
				if got.UpdatedAt.Before(before) {
					t.Error("UpdatedAt should not move backwards")
				}
			}
		})
	}
}
