package transaction_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/domain/transaction"
)

func TestNewTransaction(t *testing.T) {
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	occurredAt := time.Date(
		2026,
		time.September,
		5,
		10,
		30,
		0,
		0,
		time.FixedZone("IST", 5*60*60+30*60),
	)

	tests := []struct {
		name            string
		inputName       string
		transactionType transaction.TransactionType
		category        transaction.TransactionCategory
		fromAccountID   uuid.UUID
		toAccountID     uuid.UUID
		amount          int64
		occurredAt      time.Time
		wantName        string
		wantErr         error
	}{
		{
			name:            "valid income",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			toAccountID:     toAccountID,
			amount:          100000,
			wantName:        "Salary",
		},
		{
			name:            "valid expense",
			inputName:       "Rent",
			transactionType: transaction.TransactionTypeExpense,
			category:        transaction.TransactionCategoryHousing,
			fromAccountID:   fromAccountID,
			amount:          25000,
			wantName:        "Rent",
		},
		{
			name:            "valid transfer",
			inputName:       "Savings Transfer",
			transactionType: transaction.TransactionTypeTransfer,
			category:        transaction.TransactionCategoryTransfer,
			fromAccountID:   fromAccountID,
			toAccountID:     toAccountID,
			amount:          10000,
			wantName:        "Savings Transfer",
		},
		{
			name:            "trims whitespace",
			inputName:       "  Salary  ",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			toAccountID:     toAccountID,
			amount:          100000,
			wantName:        "Salary",
		},
		{
			name:            "empty name",
			inputName:       "",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			toAccountID:     toAccountID,
			amount:          100000,
			wantErr:         transaction.ErrInvalidName,
		},
		{
			name:            "invalid transaction type",
			inputName:       "Salary",
			transactionType: "invalid",
			category:        transaction.TransactionCategoryOther,
			amount:          100000,
			wantErr:         transaction.ErrInvalidTransactionType,
		},
		{
			name:            "invalid category",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			category:        "invalid",
			toAccountID:     toAccountID,
			amount:          100000,
			wantErr:         transaction.ErrInvalidCategory,
		},
		{
			name:            "zero amount",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			toAccountID:     toAccountID,
			amount:          0,
			wantErr:         transaction.ErrInvalidAmount,
		},
		{
			name:            "negative amount",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			toAccountID:     toAccountID,
			amount:          -1,
			wantErr:         transaction.ErrInvalidAmount,
		},
		{
			name:            "income without target",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			amount:          100000,
			wantErr:         transaction.ErrTargetAccountRequired,
		},
		{
			name:            "expense without source",
			inputName:       "Rent",
			transactionType: transaction.TransactionTypeExpense,
			category:        transaction.TransactionCategoryHousing,
			amount:          25000,
			wantErr:         transaction.ErrSourceAccountRequired,
		},
		{
			name:            "transfer without source",
			inputName:       "Transfer",
			transactionType: transaction.TransactionTypeTransfer,
			category:        transaction.TransactionCategoryTransfer,
			toAccountID:     toAccountID,
			amount:          10000,
			wantErr:         transaction.ErrInvalidAccount,
		},
		{
			name:            "transfer without target",
			inputName:       "Transfer",
			transactionType: transaction.TransactionTypeTransfer,
			category:        transaction.TransactionCategoryTransfer,
			fromAccountID:   fromAccountID,
			amount:          10000,
			wantErr:         transaction.ErrInvalidAccount,
		},
		{
			name:            "transfer to same account",
			inputName:       "Transfer",
			transactionType: transaction.TransactionTypeTransfer,
			category:        transaction.TransactionCategoryTransfer,
			fromAccountID:   fromAccountID,
			toAccountID:     fromAccountID,
			amount:          10000,
			wantErr:         transaction.ErrInvalidTransfer,
		},
		{
			name:            "normalizes occurred at",
			inputName:       "Salary",
			transactionType: transaction.TransactionTypeIncome,
			category:        transaction.TransactionCategoryOther,
			toAccountID:     toAccountID,
			amount:          100000,
			occurredAt:      occurredAt,
			wantName:        "Salary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transaction.NewTransaction(
				tt.inputName,
				tt.transactionType,
				tt.category,
				tt.fromAccountID,
				tt.toAccountID,
				uuid.Nil,
				tt.amount,
				tt.occurredAt,
				false,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"NewTransaction() error = %v, wantErr = %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Fatal("NewTransaction() returned transaction, want nil")
				}
				return
			}

			if got == nil {
				t.Fatal("NewTransaction() returned nil transaction")
			}

			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}

			if got.Type != tt.transactionType {
				t.Errorf("Type = %q, want %q", got.Type, tt.transactionType)
			}

			if got.Category != tt.category {
				t.Errorf(
					"Category = %q, want %q",
					got.Category,
					tt.category,
				)
			}

			if got.Amount != tt.amount {
				t.Errorf("Amount = %d, want %d", got.Amount, tt.amount)
			}

			if got.ID == uuid.Nil {
				t.Error("ID should not be uuid.Nil")
			}

			if got.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}

			if got.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}

			if tt.occurredAt.IsZero() {
				if got.OccurredAt.IsZero() {
					t.Error("OccurredAt should be set when input is zero")
				}
			} else {
				wantOccurredAt := tt.occurredAt.UTC().Truncate(time.Microsecond)

				if !got.OccurredAt.Equal(wantOccurredAt) {
					t.Errorf(
						"OccurredAt = %v, want %v",
						got.OccurredAt,
						wantOccurredAt,
					)
				}
			}
		})
	}
}
