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
	createFn func(context.Context, string) (*account.Account, error)
	listFn   func(context.Context) ([]*account.Account, error)
	updateFn func(
		context.Context,
		uuid.UUID,
		*string,
		*bool,
	) error
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

func (m *mockAccountService) Update(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	isArchived *bool,
) error {
	return m.updateFn(ctx, id, name, isArchived)
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
			createFn: func(
				_ context.Context,
				name string,
			) (*account.Account, error) {
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
			createFn: func(
				_ context.Context,
				_ string,
			) (*account.Account, error) {
				return nil, account.ErrAccountNameExists
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid name",
			body: `{"name":""}`,
			createFn: func(
				_ context.Context,
				_ string,
			) (*account.Account, error) {
				return nil, account.ErrInvalidName
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"name":"Savings"}`,
			createFn: func(
				_ context.Context,
				_ string,
			) (*account.Account, error) {
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

			h := handler.NewAccountHandler(
				service,
				newTestLogger(),
			)

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
		createFn: func(
			_ context.Context,
			_ string,
		) (*account.Account, error) {
			t.Fatal("Create should not be called")
			return nil, nil
		},
	}

	h := handler.NewAccountHandler(
		service,
		newTestLogger(),
	)

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

func TestAccountHandler_Patch(t *testing.T) {
	repoErr := errors.New("repository error")
	testID := uuid.New()

	tests := []struct {
		name         string
		body         string
		updateFn     func(context.Context, uuid.UUID, *string, *bool) error
		wantStatus   int
		wantName     *string
		wantArchived *bool
	}{
		{
			name: "update name",
			body: `{"name":"Updated Savings"}`,
			updateFn: func(
				_ context.Context,
				id uuid.UUID,
				name *string,
				isArchived *bool,
			) error {
				if id != testID {
					t.Errorf(
						"id = %v, want %v",
						id,
						testID,
					)
				}

				if name == nil || *name != "Updated Savings" {
					t.Errorf(
						"name = %v, want %q",
						stringPtrValue(name),
						"Updated Savings",
					)
				}

				if isArchived != nil {
					t.Errorf(
						"isArchived = %v, want nil",
						boolPtrValue(isArchived),
					)
				}

				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "archive account",
			body: `{"is_archived":true}`,
			updateFn: func(
				_ context.Context,
				id uuid.UUID,
				name *string,
				isArchived *bool,
			) error {
				if id != testID {
					t.Errorf(
						"id = %v, want %v",
						id,
						testID,
					)
				}

				if name != nil {
					t.Errorf(
						"name = %v, want nil",
						stringPtrValue(name),
					)
				}

				if isArchived == nil || !*isArchived {
					t.Errorf(
						"isArchived = %v, want true",
						boolPtrValue(isArchived),
					)
				}

				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "unarchive account",
			body: `{"is_archived":false}`,
			updateFn: func(
				_ context.Context,
				_ uuid.UUID,
				name *string,
				isArchived *bool,
			) error {
				if name != nil {
					t.Errorf(
						"name = %v, want nil",
						stringPtrValue(name),
					)
				}

				if isArchived == nil || *isArchived {
					t.Errorf(
						"isArchived = %v, want false",
						boolPtrValue(isArchived),
					)
				}

				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "update name and archive",
			body: `{"name":"Archived Savings","is_archived":true}`,
			updateFn: func(
				_ context.Context,
				id uuid.UUID,
				name *string,
				isArchived *bool,
			) error {
				if id != testID {
					t.Errorf(
						"id = %v, want %v",
						id,
						testID,
					)
				}

				if name == nil || *name != "Archived Savings" {
					t.Errorf(
						"name = %v, want %q",
						stringPtrValue(name),
						"Archived Savings",
					)
				}

				if isArchived == nil || !*isArchived {
					t.Errorf(
						"isArchived = %v, want true",
						boolPtrValue(isArchived),
					)
				}

				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "account not found",
			body: `{"name":"Savings"}`,
			updateFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ *string,
				_ *bool,
			) error {
				return account.ErrAccountNotFound
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "duplicate name",
			body: `{"name":"Existing"}`,
			updateFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ *string,
				_ *bool,
			) error {
				return account.ErrAccountNameExists
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid name",
			body: `{"name":""}`,
			updateFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ *string,
				_ *bool,
			) error {
				t.Fatal("Update should not be called")
				return nil
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repository error",
			body: `{"name":"Savings"}`,
			updateFn: func(
				_ context.Context,
				_ uuid.UUID,
				_ *string,
				_ *bool,
			) error {
				return repoErr
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockAccountService{
				updateFn: tt.updateFn,
			}

			h := handler.NewAccountHandler(
				service,
				newTestLogger(),
			)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/accounts/"+testID.String(),
				strings.NewReader(tt.body),
			)

			req = withAccountID(req, testID)

			rec := httptest.NewRecorder()

			h.Patch(rec, req)

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

func TestAccountHandler_Patch_InvalidID(t *testing.T) {
	service := &mockAccountService{
		updateFn: func(
			_ context.Context,
			_ uuid.UUID,
			_ *string,
			_ *bool,
		) error {
			t.Fatal("Update should not be called")
			return nil
		},
	}

	h := handler.NewAccountHandler(
		service,
		newTestLogger(),
	)

	req := httptest.NewRequest(
		http.MethodPatch,
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

	h.Patch(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestAccountHandler_Patch_NoFields(t *testing.T) {
	service := &mockAccountService{
		updateFn: func(
			_ context.Context,
			_ uuid.UUID,
			_ *string,
			_ *bool,
		) error {
			t.Fatal("Update should not be called")
			return nil
		},
	}

	h := handler.NewAccountHandler(
		service,
		newTestLogger(),
	)

	id := uuid.New()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/accounts/"+id.String(),
		strings.NewReader(`{}`),
	)

	req = withAccountID(req, id)

	rec := httptest.NewRecorder()

	h.Patch(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestAccountHandler_Patch_InvalidJSON(t *testing.T) {
	service := &mockAccountService{
		updateFn: func(
			_ context.Context,
			_ uuid.UUID,
			_ *string,
			_ *bool,
		) error {
			t.Fatal("Update should not be called")
			return nil
		},
	}

	h := handler.NewAccountHandler(
		service,
		newTestLogger(),
	)

	id := uuid.New()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/accounts/"+id.String(),
		strings.NewReader(`{"name":`),
	)

	req = withAccountID(req, id)

	rec := httptest.NewRecorder()

	h.Patch(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
}

func TestAccountHandler_Patch_UnknownField(t *testing.T) {
	service := &mockAccountService{
		updateFn: func(
			_ context.Context,
			_ uuid.UUID,
			_ *string,
			_ *bool,
		) error {
			t.Fatal("Update should not be called")
			return nil
		},
	}

	h := handler.NewAccountHandler(
		service,
		newTestLogger(),
	)

	id := uuid.New()

	req := httptest.NewRequest(
		http.MethodPatch,
		"/accounts/"+id.String(),
		strings.NewReader(`{"unknown":"value"}`),
	)

	req = withAccountID(req, id)

	rec := httptest.NewRecorder()

	h.Patch(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusBadRequest,
		)
	}
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
