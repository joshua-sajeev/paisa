package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/domain/account"
)

type mockAccountService struct {
	createFn     func(context.Context, string) (*account.Account, error)
	listFn       func(context.Context) ([]*account.Account, error)
	updateNameFn func(context.Context, uuid.UUID, string) error
	archiveFn    func(context.Context, uuid.UUID) error
	unarchiveFn  func(context.Context, uuid.UUID) error
}

func (m *mockAccountService) Create(
	ctx context.Context,
	name string,
) (*account.Account, error) {
	return m.createFn(ctx, name)
}

func (m *mockAccountService) List(
	ctx context.Context,
) ([]*account.Account, error) {
	return m.listFn(ctx)
}

func (m *mockAccountService) UpdateName(
	ctx context.Context,
	id uuid.UUID,
	name string,
) error {
	return m.updateNameFn(ctx, id, name)
}

func (m *mockAccountService) Archive(
	ctx context.Context,
	id uuid.UUID,
) error {
	return m.archiveFn(ctx, id)
}

func (m *mockAccountService) Unarchive(
	ctx context.Context,
	id uuid.UUID,
) error {
	return m.unarchiveFn(ctx, id)
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func withAccountID(r *http.Request, id uuid.UUID) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id.String())

	return r.WithContext(
		context.WithValue(
			r.Context(),
			chi.RouteCtxKey,
			rctx,
		),
	)
}

func TestAccountHandler_Create(t *testing.T) {
	repoErr := errors.New("repository error")

	tests := []struct {
		name       string
		body       string
		createFn   func(context.Context, string) (*account.Account, error)
		wantStatus int
		wantName   string
	}{
		{
			name: "success",
			body: `{"name":"Savings"}`,
			createFn: func(_ context.Context, name string) (*account.Account, error) {
				return &account.Account{
					ID:   uuid.New(),
					Name: name,
				}, nil
			},
			wantStatus: http.StatusCreated,
			wantName:   "Savings",
		},
		{
			name: "duplicate name",
			body: `{"name":"Savings"}`,
			createFn: func(_ context.Context, _ string) (*account.Account, error) {
				return nil, account.ErrAccountNameExists
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid name",
			body: `{"name":""}`,
			createFn: func(_ context.Context, _ string) (*account.Account, error) {
				return nil, account.ErrInvalidName
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"name":"Savings"}`,
			createFn: func(_ context.Context, _ string) (*account.Account, error) {
				return nil, repoErr
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockAccountService{
				createFn: tt.createFn,
			}

			h := handler.NewAccountHandler(service, newTestLogger())

			req := httptest.NewRequest(
				http.MethodPost,
				"/accounts",
				strings.NewReader(tt.body),
			)
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"status = %d, want %d",
					rec.Code,
					tt.wantStatus,
				)
			}

			if tt.wantName == "" {
				return
			}

			var response handler.AccountResponse

			if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if response.Name != tt.wantName {
				t.Errorf(
					"name = %q, want %q",
					response.Name,
					tt.wantName,
				)
			}
		})
	}
}

func TestAccountHandler_Create_InvalidJSON(t *testing.T) {
	service := &mockAccountService{
		createFn: func(_ context.Context, _ string) (*account.Account, error) {
			t.Fatal("Create should not be called")
			return nil, nil
		},
	}

	h := handler.NewAccountHandler(service, newTestLogger())

	req := httptest.NewRequest(
		http.MethodPost,
		"/accounts",
		strings.NewReader(`{"name":`),
	)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestAccountHandler_UpdateName(t *testing.T) {
	repoErr := errors.New("repository error")
	testID := uuid.New()

	tests := []struct {
		name         string
		accountID    uuid.UUID
		body         string
		updateNameFn func(context.Context, uuid.UUID, string) error
		wantStatus   int
	}{
		{
			name:      "success",
			accountID: testID,
			body:      `{"name":"Updated Savings"}`,
			updateNameFn: func(
				_ context.Context,
				id uuid.UUID,
				name string,
			) error {
				if id != testID {
					t.Errorf(
						"id = %v, want %v",
						id,
						testID,
					)
				}

				if name != "Updated Savings" {
					t.Errorf(
						"name = %q, want %q",
						name,
						"Updated Savings",
					)
				}

				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "account not found",
			accountID: testID,
			body:      `{"name":"Nonexistent"}`,
			updateNameFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ string,
			) error {
				return account.ErrAccountNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:      "duplicate name",
			accountID: testID,
			body:      `{"name":"Existing"}`,
			updateNameFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ string,
			) error {
				return account.ErrAccountNameExists
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:      "service error",
			accountID: testID,
			body:      `{"name":"Savings"}`,
			updateNameFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ string,
			) error {
				return repoErr
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "invalid request body",
			accountID: testID,
			body:      `{"name":`,
			updateNameFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ string,
			) error {
				t.Fatal("UpdateName should not be called")
				return nil
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "invalid name",
			accountID: testID,
			body:      `{"name":""}`,
			updateNameFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ string,
			) error {
				t.Fatal("UpdateName should not be called")
				return nil
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockAccountService{
				updateNameFn: tt.updateNameFn,
			}

			h := handler.NewAccountHandler(
				service,
				newTestLogger(),
			)

			req := httptest.NewRequest(
				http.MethodPut,
				"/accounts/"+tt.accountID.String(),
				strings.NewReader(tt.body),
			)

			req = withAccountID(req, tt.accountID)

			rec := httptest.NewRecorder()

			h.UpdateName(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"status = %d, want %d",
					rec.Code,
					tt.wantStatus,
				)
			}
		})
	}
}

func TestAccountHandler_Update_InvalidID(t *testing.T) {
	service := &mockAccountService{
		updateNameFn: func(
			_ context.Context,
			_ uuid.UUID,
			_ string,
		) error {
			t.Fatal("UpdateName should not be called")
			return nil
		},
	}

	h := handler.NewAccountHandler(
		service,
		newTestLogger(),
	)

	req := httptest.NewRequest(
		http.MethodPut,
		"/accounts/not-a-uuid",
		strings.NewReader(`{"name":"Updated"}`),
	)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")

	req = req.WithContext(
		context.WithValue(
			req.Context(),
			chi.RouteCtxKey,
			rctx,
		),
	)

	rec := httptest.NewRecorder()

	h.UpdateName(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}
