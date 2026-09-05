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

	if got.IsArchived != want.IsArchived {
		t.Errorf("IsArchived = %v, want %v", got.IsArchived, want.IsArchived)
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
		truncateAccountsTable(t, ctx, db)
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
		truncateAccountsTable(t, ctx, db)
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
		truncateAccountsTable(t, ctx, db)
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
		truncateAccountsTable(t, ctx, db)
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

func TestAccountUpdateName(t *testing.T) {
	t.Cleanup(func() {
		truncateAccountsTable(t, ctx, db)
	})

	tests := []struct {
		name    string
		newName string
	}{
		{
			name:    "valid name",
			newName: "Updated Name",
		},
		{
			name:    "empty name",
			newName: "",
		},
		{
			name:    "unicode name",
			newName: "新しい名前",
		},
		{
			name:    "long name",
			newName: "VeryLongUpdatedNameWithManyCharacters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAccount("Original")

			if err := accountRepo.Create(ctx, a); err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			originalCreatedAt := a.CreatedAt
			originalUpdatedAt := a.UpdatedAt

			time.Sleep(10 * time.Millisecond)

			if err := accountRepo.UpdateName(ctx, a.ID, tt.newName); err != nil {
				t.Fatalf("UpdateName() error = %v", err)
			}

			got := queryAccount(t, a.ID)

			if got.Name != tt.newName {
				t.Errorf(
					"Name = %q, want %q",
					got.Name,
					tt.newName,
				)
			}

			if got.IsArchived != a.IsArchived {
				t.Errorf(
					"IsArchived changed unexpectedly: got=%v, want=%v",
					got.IsArchived,
					a.IsArchived,
				)
			}

			if !got.CreatedAt.Equal(originalCreatedAt) {
				t.Errorf(
					"CreatedAt changed unexpectedly: got=%v, want=%v",
					got.CreatedAt,
					originalCreatedAt,
				)
			}

			if !got.UpdatedAt.After(originalUpdatedAt) {
				t.Errorf(
					"UpdatedAt not updated: orig=%v, got=%v",
					originalUpdatedAt,
					got.UpdatedAt,
				)
			}
		})
	}

	t.Run("non-existent account", func(t *testing.T) {
		id := uuid.New()

		err := accountRepo.UpdateName(ctx, id, "Name")

		if !errors.Is(err, account.ErrAccountNotFound) {
			t.Errorf(
				"UpdateName() error = %v, want account.ErrAccountNotFound",
				err,
			)
		}
	})
}

func TestAccountSetArchived(t *testing.T) {
	t.Cleanup(func() {
		truncateAccountsTable(t, ctx, db)
	})

	a := newTestAccount("Lifecycle")

	if err := accountRepo.Create(ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	original := queryAccount(t, a.ID)

	time.Sleep(10 * time.Millisecond)

	if err := accountRepo.SetArchived(ctx, a.ID, true); err != nil {
		t.Fatalf("SetArchived(true) error = %v", err)
	}

	got := queryAccount(t, a.ID)

	if !got.IsArchived {
		t.Error("IsArchived = false, want true")
	}

	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf(
			"CreatedAt changed: got=%v, want=%v",
			got.CreatedAt,
			original.CreatedAt,
		)
	}

	if !got.UpdatedAt.After(original.UpdatedAt) {
		t.Errorf(
			"UpdatedAt not updated: orig=%v, got=%v",
			original.UpdatedAt,
			got.UpdatedAt,
		)
	}

	beforeUnarchive := got.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	if err := accountRepo.SetArchived(ctx, a.ID, false); err != nil {
		t.Fatalf("SetArchived(false) error = %v", err)
	}

	got = queryAccount(t, a.ID)

	if got.IsArchived {
		t.Error("IsArchived = true, want false")
	}

	if !got.UpdatedAt.After(beforeUnarchive) {
		t.Errorf(
			"UpdatedAt not updated on unarchive: before=%v, got=%v",
			beforeUnarchive,
			got.UpdatedAt,
		)
	}
}

func TestAccountSetArchived_NotFound(t *testing.T) {
	t.Cleanup(func() {
		truncateAccountsTable(t, ctx, db)
	})

	err := accountRepo.SetArchived(ctx, uuid.New(), true)

	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf(
			"SetArchived() error = %v, want account.ErrAccountNotFound",
			err,
		)
	}
}

func TestConcurrentSetArchived(t *testing.T) {
	t.Cleanup(func() {
		truncateAccountsTable(t, ctx, db)
	})

	const numOps = 10

	ids := make([]uuid.UUID, numOps)

	for i := range numOps {
		a := newTestAccount(
			fmt.Sprintf("Concurrent Archive %d", i),
		)

		ids[i] = a.ID

		if err := accountRepo.Create(ctx, a); err != nil {
			t.Fatalf("setup Create() error: %v", err)
		}
	}

	errChan := make(chan error, numOps)

	for i := range numOps {
		go func(idx int) {
			errChan <- accountRepo.SetArchived(
				ctx,
				ids[idx],
				true,
			)
		}(i)
	}

	for range numOps {
		if err := <-errChan; err != nil {
			t.Errorf(
				"SetArchived(true) error: %v",
				err,
			)
		}
	}

	for _, id := range ids {
		got := queryAccount(t, id)

		if !got.IsArchived {
			t.Errorf(
				"account %v IsArchived = false, want true",
				id,
			)
		}
	}

	for i := range numOps {
		go func(idx int) {
			errChan <- accountRepo.SetArchived(
				ctx,
				ids[idx],
				false,
			)
		}(i)
	}

	for range numOps {
		if err := <-errChan; err != nil {
			t.Errorf(
				"SetArchived(false) error: %v",
				err,
			)
		}
	}

	for _, id := range ids {
		got := queryAccount(t, id)

		if got.IsArchived {
			t.Errorf(
				"account %v IsArchived = true, want false",
				id,
			)
		}
	}
}
