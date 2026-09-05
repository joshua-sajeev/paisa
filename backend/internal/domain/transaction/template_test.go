package transaction_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
)

func TestNewTemplate(t *testing.T) {
	accountID := uuid.New()
	jarID := uuid.New()

	amount := int64(100000)
	zeroAmount := int64(0)
	negativeAmount := int64(-1)

	tests := []struct {
		name            string
		inputName       string
		transactionType transaction.TransactionType
		fromAccountID   uuid.UUID
		toAccountID     uuid.UUID
		jarID           uuid.UUID
		amount          *int64
		wantName        string
		wantAmount      *int64
		wantErr         error
	}{
		{
			name:            "valid template with amount",
			inputName:       "Monthly Salary",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			jarID:           jarID,
			amount:          &amount,
			wantName:        "Monthly Salary",
			wantAmount:      &amount,
		},
		{
			name:            "valid template without amount",
			inputName:       "Monthly Salary",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			jarID:           jarID,
			amount:          nil,
			wantName:        "Monthly Salary",
			wantAmount:      nil,
		},
		{
			name:            "trims whitespace",
			inputName:       "  Monthly Salary  ",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			amount:          nil,
			wantName:        "Monthly Salary",
			wantAmount:      nil,
		},
		{
			name:            "empty name",
			inputName:       "",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			amount:          nil,
			wantErr:         transaction.ErrTemplateInvalidName,
		},
		{
			name:            "whitespace only name",
			inputName:       "   ",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			amount:          nil,
			wantErr:         transaction.ErrTemplateInvalidName,
		},
		{
			name:            "invalid transaction type",
			inputName:       "Salary",
			transactionType: "invalid",
			amount:          nil,
			wantErr:         transaction.ErrInvalidTransactionType,
		},
		{
			name:            "zero amount",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			amount:          &zeroAmount,
			wantErr:         transaction.ErrTemplateInvalidAmount,
		},
		{
			name:            "negative amount",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			toAccountID:     accountID,
			amount:          &negativeAmount,
			wantErr:         transaction.ErrTemplateInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transaction.NewTemplate(
				tt.inputName,
				tt.transactionType,
				tt.fromAccountID,
				tt.toAccountID,
				tt.jarID,
				tt.amount,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"NewTemplate() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Fatal("NewTemplate() returned template, want nil")
				}
				return
			}

			if got == nil {
				t.Fatal("NewTemplate() returned nil template")
			}

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}

			if got.Type != tt.transactionType {
				t.Errorf("Type = %q, want %q", got.Type, tt.transactionType)
			}

			if got.Amount == nil && tt.wantAmount != nil {
				t.Fatal("Amount = nil, want non-nil")
			}

			if got.Amount != nil && tt.wantAmount == nil {
				t.Fatal("Amount is non-nil, want nil")
			}

			if got.Amount != nil && tt.wantAmount != nil {
				if *got.Amount != *tt.wantAmount {
					t.Errorf(
						"Amount = %d, want %d",
						*got.Amount,
						*tt.wantAmount,
					)
				}
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.IsArchived {
				t.Error("new template should not be archived")
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
