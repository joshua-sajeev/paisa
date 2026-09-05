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
	createFn      func(context.Context, *account.Account) error
	listFn        func(context.Context) ([]*account.Account, error)
	updateNameFn  func(ctx context.Context, id uuid.UUID, name string) error
	setArchivedFn func(context.Context, uuid.UUID, bool) error
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

func (m *mockAccountRepository) UpdateName(
	ctx context.Context,
	id uuid.UUID,
	name string,
) error {
	return m.updateNameFn(ctx, id, name)
}

func (m *mockAccountRepository) SetArchived(
	ctx context.Context,
	id uuid.UUID,
	archived bool,
) error {
	return m.setArchivedFn(ctx, id, archived)
}

func newTestAccountService(repo *mockAccountRepository) *application.AccountService {
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

func TestAccountService_UpdateName(t *testing.T) {
	repoErr := errors.New("repository error")

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{
			name: "success",
		},
		{
			name:    "repository error",
			repoErr: repoErr,
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			name := "Savings"

			var gotID uuid.UUID
			var gotName string

			repo := &mockAccountRepository{
				updateNameFn: func(
					_ context.Context,
					id uuid.UUID,
					name string,
				) error {
					gotID = id
					gotName = name
					return tt.repoErr
				},
			}

			service := newTestAccountService(repo)

			err := service.UpdateName(
				context.Background(),
				id,
				name,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"UpdateName() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if gotID != id {
				t.Errorf(
					"repository received id = %v, want %v",
					gotID,
					id,
				)
			}

			if gotName != name {
				t.Errorf(
					"repository received name = %q, want %q",
					gotName,
					name,
				)
			}
		})
	}
}

func TestAccountService_Archive(t *testing.T) {
	id := uuid.New()
	repoErr := errors.New("repository error")

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{
			name: "success",
		},
		{
			name:    "repository error",
			repoErr: repoErr,
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID uuid.UUID
			var gotArchived bool

			repo := &mockAccountRepository{
				setArchivedFn: func(
					_ context.Context,
					id uuid.UUID,
					archived bool,
				) error {
					gotID = id
					gotArchived = archived

					return tt.repoErr
				},
			}

			service := newTestAccountService(repo)

			err := service.Archive(
				context.Background(),
				id,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"Archive() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if gotID != id {
				t.Errorf(
					"repository id = %v, want %v",
					gotID,
					id,
				)
			}

			if !gotArchived {
				t.Error(
					"repository archived = false, want true",
				)
			}
		})
	}
}

func TestAccountService_Unarchive(t *testing.T) {
	id := uuid.New()
	repoErr := errors.New("repository error")

	tests := []struct {
		name    string
		repoErr error
		wantErr error
	}{
		{
			name: "success",
		},
		{
			name:    "repository error",
			repoErr: repoErr,
			wantErr: repoErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID uuid.UUID
			var gotArchived bool

			repo := &mockAccountRepository{
				setArchivedFn: func(
					_ context.Context,
					id uuid.UUID,
					archived bool,
				) error {
					gotID = id
					gotArchived = archived

					return tt.repoErr
				},
			}

			service := newTestAccountService(repo)

			err := service.Unarchive(
				context.Background(),
				id,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"Unarchive() error = %v, want %v",
					err,
					tt.wantErr,
				)
			}

			if gotID != id {
				t.Errorf(
					"repository id = %v, want %v",
					gotID,
					id,
				)
			}

			if gotArchived {
				t.Error(
					"repository archived = true, want false",
				)
			}
		})
	}
}
