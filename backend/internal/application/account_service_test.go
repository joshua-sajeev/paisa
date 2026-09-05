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
	createFn func(context.Context, *account.Account) error
	listFn   func(context.Context) ([]*account.Account, error)
	updateFn func(
		context.Context,
		uuid.UUID,
		*string,
		*bool,
	) error
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

func (m *mockAccountRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	isArchived *bool,
) error {
	return m.updateFn(ctx, id, name, isArchived)
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
		updateName     *string
		updateArchived *bool
		repoErr        error
		wantErr        error
	}{
		{
			name: "update name only",
			updateName: func() *string {
				name := "Savings"
				return &name
			}(),
		},
		{
			name: "archive only",
			updateArchived: func() *bool {
				archived := true
				return &archived
			}(),
		},
		{
			name: "unarchive only",
			updateArchived: func() *bool {
				archived := false
				return &archived
			}(),
		},
		{
			name: "update name and archive",
			updateName: func() *string {
				name := "Archived Savings"
				return &name
			}(),
			updateArchived: func() *bool {
				archived := true
				return &archived
			}(),
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

			var gotID uuid.UUID
			var gotName *string
			var gotArchived *bool

			repo := &mockAccountRepository{
				updateFn: func(
					_ context.Context,
					id uuid.UUID,
					name *string,
					isArchived *bool,
				) error {
					gotID = id
					gotName = name
					gotArchived = isArchived

					return tt.repoErr
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

			if gotID != id {
				t.Errorf(
					"repository received id = %v, want %v",
					gotID,
					id,
				)
			}

			if !equalStringPtr(gotName, tt.updateName) {
				t.Errorf(
					"repository received name = %v, want %v",
					stringPtrValue(gotName),
					stringPtrValue(tt.updateName),
				)
			}

			if !equalBoolPtr(gotArchived, tt.updateArchived) {
				t.Errorf(
					"repository received archived = %v, want %v",
					boolPtrValue(gotArchived),
					boolPtrValue(tt.updateArchived),
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
