package application_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/application"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

type mockAccountRepository struct {
	createFn        func(context.Context, *account.Account) error
	listFn          func(context.Context) ([]*account.Account, error)
	findByIDFn      func(context.Context, uuid.UUID) (*account.Account, error)
	saveFn          func(context.Context, *account.Account) error
	adjustBalanceFn func(context.Context, uuid.UUID, int64) error
}

func (m *mockAccountRepository) Create(
	ctx context.Context,
	a *account.Account,
) error {
	return m.createFn(ctx, a)
}

func (m *mockAccountRepository) List(
	ctx context.Context,
) ([]*account.Account, error) {
	return m.listFn(ctx)
}

func (m *mockAccountRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*account.Account, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockAccountRepository) Save(
	ctx context.Context,
	a *account.Account,
) error {
	return m.saveFn(ctx, a)
}

func (m *mockAccountRepository) AdjustBalance(
	ctx context.Context,
	id uuid.UUID,
	delta int64,
) error {
	if m.adjustBalanceFn == nil {
		return nil
	}

	return m.adjustBalanceFn(ctx, id, delta)
}

func newTestAccountService(
	repo *mockAccountRepository,
) *application.AccountService {
	logger := slog.New(slog.DiscardHandler)

	return application.NewAccountService(repo, logger)
}

func TestAccountService_Create(t *testing.T) {
	repoErr := errors.New("repository error")

	tests := []struct {
		name        string
		accountName string
		repoErr     error
		wantErr     error
		wantName    string
	}{
		{
			name:        "success",
			accountName: "Savings",
			wantName:    "Savings",
		},
		{
			name:        "duplicate name",
			accountName: "Savings",
			repoErr:     account.ErrAccountNameExists,
			wantErr:     account.ErrAccountNameExists,
		},
		{
			name:        "repository error",
			accountName: "Savings",
			repoErr:     repoErr,
			wantErr:     repoErr,
		},
		{
			name:        "invalid name",
			accountName: "",
			wantErr:     account.ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountRepository{
				createFn: func(
					_ context.Context,
					a *account.Account,
				) error {
					return tt.repoErr
				},
			}

			service := newTestAccountService(repo)

			got, err := service.Create(
				context.Background(),
				tt.accountName,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"Create() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf(
						"Create() account = %v, want nil",
						got,
					)
				}

				return
			}

			if got == nil {
				t.Fatal("Create() returned nil account")
			}

			if got.Name != tt.wantName {
				t.Errorf(
					"Create() name = %q, want %q",
					got.Name,
					tt.wantName,
				)
			}

			if got.ID == uuid.Nil {
				t.Error("Create() returned account with nil ID")
			}
		})
	}
}

func TestAccountService_List(t *testing.T) {
	repoErr := errors.New("repository error")

	accounts := []*account.Account{
		{
			ID:   uuid.New(),
			Name: "Savings",
		},
		{
			ID:   uuid.New(),
			Name: "Checking",
		},
	}

	tests := []struct {
		name     string
		accounts []*account.Account
		repoErr  error
		wantLen  int
		wantErr  error
	}{
		{
			name:     "success",
			accounts: accounts,
			wantLen:  2,
		},
		{
			name:     "empty result",
			accounts: []*account.Account{},
			wantLen:  0,
		},
		{
			name:    "repository error",
			repoErr: repoErr,
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAccountRepository{
				listFn: func(
					_ context.Context,
				) ([]*account.Account, error) {
					return tt.accounts, tt.repoErr
				},
			}

			service := newTestAccountService(repo)

			got, err := service.List(context.Background())

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"List() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if tt.wantErr != nil {
				if got != nil {
					t.Errorf(
						"List() = %v, want nil",
						got,
					)
				}

				return
			}

			if len(got) != tt.wantLen {
				t.Errorf(
					"List() length = %d, want %d",
					len(got),
					tt.wantLen,
				)
			}
		})
	}
}

func TestAccountService_Update(t *testing.T) {
	repoErr := errors.New("repository error")

	tests := []struct {
		name           string
		initial        *account.Account
		updateName     *string
		updateArchived *bool
		findErr        error
		saveErr        error
		wantErr        error
		wantName       string
		wantArchived   bool
		wantSave       bool
	}{
		{
			name: "update name",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Checking",
				IsArchived: false,
			},
			updateName: func() *string {
				name := "Savings"
				return &name
			}(),
			wantName:     "Savings",
			wantArchived: false,
			wantSave:     true,
		},
		{
			name: "archive",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Savings",
				IsArchived: false,
			},
			updateArchived: func() *bool {
				archived := true
				return &archived
			}(),
			wantName:     "Savings",
			wantArchived: true,
			wantSave:     true,
		},
		{
			name: "unarchive",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Savings",
				IsArchived: true,
			},
			updateArchived: func() *bool {
				archived := false
				return &archived
			}(),
			wantName:     "Savings",
			wantArchived: false,
			wantSave:     true,
		},
		{
			name: "update name and archive",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Savings",
				IsArchived: false,
			},
			updateName: func() *string {
				name := "Archived Savings"
				return &name
			}(),
			updateArchived: func() *bool {
				archived := true
				return &archived
			}(),
			wantName:     "Archived Savings",
			wantArchived: true,
			wantSave:     true,
		},
		{
			name: "already archived",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Savings",
				IsArchived: true,
			},
			updateArchived: func() *bool {
				archived := true
				return &archived
			}(),
			wantName:     "Savings",
			wantArchived: true,
			wantSave:     false,
		},
		{
			name: "already active",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Savings",
				IsArchived: false,
			},
			updateArchived: func() *bool {
				archived := false
				return &archived
			}(),
			wantName:     "Savings",
			wantArchived: false,
			wantSave:     false,
		},
		{
			name:     "account not found",
			initial:  nil,
			findErr:  account.ErrAccountNotFound,
			wantErr:  account.ErrAccountNotFound,
			wantSave: false,
		},
		{
			name:     "find repository error",
			initial:  nil,
			findErr:  repoErr,
			wantErr:  repoErr,
			wantSave: false,
		},
		{
			name: "save repository error",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Checking",
				IsArchived: false,
			},
			updateName: func() *string {
				name := "Savings"
				return &name
			}(),
			saveErr:      repoErr,
			wantErr:      repoErr,
			wantName:     "Savings",
			wantArchived: false,
			wantSave:     true,
		},
		{
			name: "invalid name",
			initial: &account.Account{
				ID:         uuid.New(),
				Name:       "Checking",
				IsArchived: false,
			},
			updateName: func() *string {
				name := ""
				return &name
			}(),
			wantErr:      account.ErrInvalidName,
			wantName:     "Checking",
			wantArchived: false,
			wantSave:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()

			if tt.initial != nil {
				tt.initial.ID = id
			}

			var findID uuid.UUID
			var saved *account.Account
			saveCalled := false

			repo := &mockAccountRepository{
				findByIDFn: func(
					_ context.Context,
					gotID uuid.UUID,
				) (*account.Account, error) {
					findID = gotID

					if tt.findErr != nil {
						return nil, tt.findErr
					}

					return tt.initial, nil
				},
				saveFn: func(
					_ context.Context,
					a *account.Account,
				) error {
					saveCalled = true
					saved = a

					return tt.saveErr
				},
			}

			service := newTestAccountService(repo)

			err := service.Update(
				context.Background(),
				id,
				tt.updateName,
				tt.updateArchived,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"Update() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if findID != id {
				t.Errorf(
					"FindByID() id = %v, want %v",
					findID,
					id,
				)
			}

			if saveCalled != tt.wantSave {
				t.Errorf(
					"Save() called = %v, want %v",
					saveCalled,
					tt.wantSave,
				)
			}

			if !tt.wantSave {
				return
			}

			if saved == nil {
				t.Fatal("Save() received nil account")
			}

			if saved.Name != tt.wantName {
				t.Errorf(
					"saved Name = %q, want %q",
					saved.Name,
					tt.wantName,
				)
			}

			if saved.IsArchived != tt.wantArchived {
				t.Errorf(
					"saved IsArchived = %v, want %v",
					saved.IsArchived,
					tt.wantArchived,
				)
			}
		})
	}
}

func equalStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func equalBoolPtr(a, b *bool) bool {
	if a == nil || b == nil {
		return a == b
	}

	return *a == *b
}

func stringPtrValue(value *string) string {
	if value == nil {
		return "<nil>"
	}

	return *value
}

func boolPtrValue(value *bool) string {
	if value == nil {
		return "<nil>"
	}

	if *value {
		return "true"
	}

	return "false"
}
