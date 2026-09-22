package postgres_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

func newTestAccount(name string) *account.Account {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &account.Account{
		ID:         uuid.New(),
		Name:       name + " " + uuid.NewString(),
		IconKey:    "bank",
		IsArchived: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func assertAccount(t *testing.T, got, want *account.Account) {
	t.Helper()

	if got.ID != want.ID {
		t.Errorf("ID = %v, want %v", got.ID, want.ID)
	}

	if got.Name != want.Name {
		t.Errorf("Name = %q, want %q", got.Name, want.Name)
	}

	if got.IconKey != want.IconKey {
		t.Errorf("IconKey = %q, want %q", got.IconKey, want.IconKey)
	}

	if got.Balance != want.Balance {
		t.Errorf("Balance = %d, want %d", got.Balance, want.Balance)
	}

	if got.IsArchived != want.IsArchived {
		t.Errorf(
			"IsArchived = %v, want %v",
			got.IsArchived,
			want.IsArchived,
		)
	}

	if got.CreatedAt.IsZero() {
		t.Errorf("CreatedAt = %v, want non-zero time", got.CreatedAt)
	}

	if got.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt = %v, want non-zero time", got.UpdatedAt)
	}
}

// queryAccount reads an account directly from PostgreSQL.
// It is intentionally a test helper, not part of the repository port.
func queryAccount(t *testing.T, id uuid.UUID) *account.Account {
	t.Helper()

	var a account.Account

	err := db.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			icon_key,
			balance,
			is_archived,
			created_at,
			updated_at
		FROM accounts
		WHERE id = $1
		`,
		id,
	).Scan(
		&a.ID,
		&a.Name,
		&a.IconKey,
		&a.Balance,
		&a.IsArchived,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("query account: %v", err)
	}

	return &a
}

func TestAccountCreate(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	tests := []struct {
		name string
		acc  *account.Account
	}{
		{"basic account", newTestAccount("Savings")},
		{"checking account", newTestAccount("Checking")},
		{"investment account", newTestAccount("Investment")},
		{"special chars", newTestAccount(`Account's "Savings"`)},
		{"unicode", newTestAccount("日本語アカウント")},
		{"long name", newTestAccount("VeryLongAccountNameWithManyCharactersForTesting")},
		{
			name: "initial balance",
			acc: func() *account.Account {
				acc := newTestAccount("Opening Balance")
				acc.Balance = 125000
				return acc
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := accountRepo.Create(ctx, tt.acc); err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			got := queryAccount(t, tt.acc.ID)

			assertAccount(t, got, tt.acc)
		})
	}
}

func TestAccountCreate_DuplicateName(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	first := newTestAccount("Savings")

	if err := accountRepo.Create(ctx, first); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	second := newTestAccount("Other")
	second.Name = first.Name

	err := accountRepo.Create(ctx, second)

	if !errors.Is(err, account.ErrAccountNameExists) {
		t.Errorf(
			"Create() error = %v, want account.ErrAccountNameExists",
			err,
		)
	}
}

func TestAccountCreate_ConcurrentDuplicateName(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	const numOps = 10
	const name = "Concurrent Savings"

	errChan := make(chan error, numOps)

	for range numOps {
		go func() {
			a := newTestAccount("Other")
			a.Name = name

			errChan <- accountRepo.Create(ctx, a)
		}()
	}

	var successCount int
	var duplicateCount int

	for range numOps {
		err := <-errChan

		switch {
		case err == nil:
			successCount++

		case errors.Is(err, account.ErrAccountNameExists):
			duplicateCount++

		default:
			t.Errorf("unexpected Create() error = %v", err)
		}
	}

	if successCount != 1 {
		t.Errorf("successful creates = %d, want 1", successCount)
	}

	if duplicateCount != numOps-1 {
		t.Errorf(
			"duplicate errors = %d, want %d",
			duplicateCount,
			numOps-1,
		)
	}
}

func TestAccountList(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	account1 := newTestAccount("Checking")
	account1.CreatedAt = time.Date(
		2026, 1, 1, 0, 0, 0, 0, time.UTC,
	)
	account1.UpdatedAt = account1.CreatedAt

	account2 := newTestAccount("Savings")
	account2.CreatedAt = time.Date(
		2026, 1, 2, 0, 0, 0, 0, time.UTC,
	)
	account2.UpdatedAt = account2.CreatedAt

	archived := newTestAccount("Archived Account")
	archived.IsArchived = true
	archived.CreatedAt = time.Date(
		2026, 1, 3, 0, 0, 0, 0, time.UTC,
	)
	archived.UpdatedAt = archived.CreatedAt

	for _, a := range []*account.Account{
		account1,
		account2,
		archived,
	} {
		if err := accountRepo.Create(ctx, a); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	got, err := accountRepo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("List() returned %d accounts, want 3", len(got))
	}

	if got[0].ID != archived.ID {
		t.Errorf(
			"first account ID = %v, want %v",
			got[0].ID,
			archived.ID,
		)
	}

	if got[1].ID != account2.ID {
		t.Errorf(
			"second account ID = %v, want %v",
			got[1].ID,
			account2.ID,
		)
	}

	if got[2].ID != account1.ID {
		t.Errorf(
			"third account ID = %v, want %v",
			got[2].ID,
			account1.ID,
		)
	}

	if !got[0].IsArchived {
		t.Error("first account IsArchived = false, want true")
	}
}

func TestAccountFindByID(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	acc := newTestAccount("Savings")

	if err := accountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := accountRepo.FindByID(ctx, acc.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	assertAccount(t, got, acc)
}

func TestAccountFindByID_NotFound(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	_, err := accountRepo.FindByID(ctx, uuid.New())

	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf(
			"FindByID() error = %v, want %v",
			err,
			account.ErrAccountNotFound,
		)
	}
}

func TestAccountSave(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	tests := []struct {
		name         string
		setup        func(t *testing.T) *account.Account
		update       func(*account.Account)
		wantName     string
		wantArchived bool
		wantIcon     string
		wantErr      error
	}{
		{
			name: "update name",
			setup: func(t *testing.T) *account.Account {
				acc := newTestAccount("Original")

				if err := accountRepo.Create(ctx, acc); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				return acc
			},
			update: func(a *account.Account) {
				a.Name = "Updated Name"
				a.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
			},
			wantName:     "Updated Name",
			wantArchived: false,
		},
		{
			name: "update icon",
			setup: func(t *testing.T) *account.Account {
				acc := newTestAccount("Original")

				if err := accountRepo.Create(ctx, acc); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				return acc
			},
			update: func(a *account.Account) {
				a.UpdateIcon("credit-card")
			},
			wantName:     "",
			wantArchived: false,
			wantIcon:     "credit-card",
		},
		{
			name: "archive",
			setup: func(t *testing.T) *account.Account {
				acc := newTestAccount("Original")

				if err := accountRepo.Create(ctx, acc); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				return acc
			},
			update: func(a *account.Account) {
				a.IsArchived = true
				a.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
			},
			wantName:     "", // Keep existing name.
			wantArchived: true,
		},
		{
			name: "unarchive",
			setup: func(t *testing.T) *account.Account {
				acc := newTestAccount("Original")
				acc.IsArchived = true

				if err := accountRepo.Create(ctx, acc); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				return acc
			},
			update: func(a *account.Account) {
				a.IsArchived = false
				a.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
			},
			wantName:     "",
			wantArchived: false,
		},
		{
			name: "update name and archive",
			setup: func(t *testing.T) *account.Account {
				acc := newTestAccount("Original")

				if err := accountRepo.Create(ctx, acc); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				return acc
			},
			update: func(a *account.Account) {
				a.Name = "Archived Savings"
				a.IsArchived = true
				a.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
			},
			wantName:     "Archived Savings",
			wantArchived: true,
		},
		{
			name: "unicode name",
			setup: func(t *testing.T) *account.Account {
				acc := newTestAccount("Original")

				if err := accountRepo.Create(ctx, acc); err != nil {
					t.Fatalf("Create() error = %v", err)
				}

				return acc
			},
			update: func(a *account.Account) {
				a.Name = "新しい名前"
				a.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
			},
			wantName:     "新しい名前",
			wantArchived: false,
		},
		{
			name: "duplicate name",
			setup: func(t *testing.T) *account.Account {
				first := newTestAccount("First")
				second := newTestAccount("Second")

				if err := accountRepo.Create(ctx, first); err != nil {
					t.Fatalf("first Create() error = %v", err)
				}

				if err := accountRepo.Create(ctx, second); err != nil {
					t.Fatalf("second Create() error = %v", err)
				}

				second.Name = first.Name
				second.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

				return second
			},
			wantErr: account.ErrAccountNameExists,
		},
		{
			name: "not found",
			setup: func(t *testing.T) *account.Account {
				return newTestAccount("Missing")
			},
			wantErr: account.ErrAccountNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.setup(t)

			originalName := acc.Name
			originalCreatedAt := acc.CreatedAt

			if tt.update != nil {
				tt.update(acc)
			}

			err := accountRepo.Save(ctx, acc)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"Save() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				return
			}

			got := queryAccount(t, acc.ID)

			wantName := tt.wantName
			if wantName == "" {
				wantName = originalName
			}

			if got.Name != wantName {
				t.Errorf(
					"Name = %q, want %q",
					got.Name,
					wantName,
				)
			}

			if got.IsArchived != tt.wantArchived {
				t.Errorf(
					"IsArchived = %v, want %v",
					got.IsArchived,
					tt.wantArchived,
				)
			}

			wantIcon := tt.wantIcon
			if wantIcon == "" {
				wantIcon = acc.IconKey
			}

			if got.IconKey != wantIcon {
				t.Errorf(
					"IconKey = %q, want %q",
					got.IconKey,
					wantIcon,
				)
			}

			if !got.CreatedAt.Equal(originalCreatedAt) {
				t.Errorf(
					"CreatedAt changed unexpectedly: got=%v, want=%v",
					got.CreatedAt,
					originalCreatedAt,
				)
			}

			if !got.UpdatedAt.Equal(acc.UpdatedAt) {
				t.Errorf(
					"UpdatedAt = %v, want %v",
					got.UpdatedAt,
					acc.UpdatedAt,
				)
			}
		})
	}
}

func TestAccountSave_DoesNotOverwriteBalance(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	acc := newTestAccount("Checking")
	acc.Balance = 1000

	if err := accountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := accountRepo.AdjustBalance(ctx, acc.ID, 500); err != nil {
		t.Fatalf("AdjustBalance() error = %v", err)
	}

	acc.Name = "Renamed Checking"
	acc.Balance = 1
	acc.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

	if err := accountRepo.Save(ctx, acc); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got := queryAccount(t, acc.ID)

	if got.Name != "Renamed Checking" {
		t.Errorf("Name = %q, want %q", got.Name, "Renamed Checking")
	}

	if got.Balance != 1500 {
		t.Errorf("Balance = %d, want %d", got.Balance, 1500)
	}
}

func TestAccountAdjustBalance(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	acc := newTestAccount("Checking")
	acc.Balance = 1000

	if err := accountRepo.Create(ctx, acc); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := accountRepo.AdjustBalance(ctx, acc.ID, 2500); err != nil {
		t.Fatalf("AdjustBalance() error = %v", err)
	}

	if err := accountRepo.AdjustBalance(ctx, acc.ID, -750); err != nil {
		t.Fatalf("AdjustBalance() error = %v", err)
	}

	got := queryAccount(t, acc.ID)

	if got.Balance != 2750 {
		t.Errorf("Balance = %d, want %d", got.Balance, 2750)
	}
}

func TestAccountAdjustBalance_NotFound(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	err := accountRepo.AdjustBalance(ctx, uuid.New(), 1000)

	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf(
			"AdjustBalance() error = %v, want %v",
			err,
			account.ErrAccountNotFound,
		)
	}
}

func TestAccountSave_Concurrent(t *testing.T) {
	t.Cleanup(func() {
		truncateTables(t, ctx, db)
	})

	const numOps = 10

	accounts := make([]*account.Account, numOps)

	for i := range numOps {
		acc := newTestAccount(
			fmt.Sprintf("Concurrent Account %d", i),
		)

		if err := accountRepo.Create(ctx, acc); err != nil {
			t.Fatalf("setup Create() error: %v", err)
		}

		accounts[i] = acc
	}

	errChan := make(chan error, numOps)

	for i := range numOps {
		go func(idx int) {
			acc := accounts[idx]

			acc.IsArchived = true
			acc.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)

			errChan <- accountRepo.Save(ctx, acc)
		}(i)
	}

	for range numOps {
		if err := <-errChan; err != nil {
			t.Errorf("Save() error: %v", err)
		}
	}

	for _, acc := range accounts {
		got := queryAccount(t, acc.ID)

		if !got.IsArchived {
			t.Errorf(
				"account %v IsArchived = false, want true",
				acc.ID,
			)
		}
	}
}
