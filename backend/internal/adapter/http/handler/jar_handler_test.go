package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joshu-sajeev/paisa/internal/adapter/http/handler"
	"github.com/joshu-sajeev/paisa/internal/domain/jar"
)

type MockJarService struct {
	CreateCalls            int
	ListCalls              int
	UpdateCalls            int
	UpdateAllocationsCalls int

	CreateError            error
	ListError              error
	UpdateError            error
	UpdateAllocationsError error

	CreatedJar         *jar.Jar
	Jars               []*jar.Jar
	UpdatedAllocations []jar.JarAllocationUpdate
}

func (m *MockJarService) Create(
	ctx context.Context,
	name string,
	allocationType jar.AllocationType,
	allocationValue int64,
) (*jar.Jar, error) {
	m.CreateCalls++

	if m.CreateError != nil {
		return nil, m.CreateError
	}

	if m.CreatedJar != nil {
		return m.CreatedJar, nil
	}

	now := time.Now().UTC()

	return &jar.Jar{
		ID:              uuid.New(),
		Name:            name,
		AllocationType:  allocationType,
		AllocationValue: allocationValue,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (m *MockJarService) List(ctx context.Context) ([]*jar.Jar, error) {
	m.ListCalls++

	if m.ListError != nil {
		return nil, m.ListError
	}

	return m.Jars, nil
}

func (m *MockJarService) Update(
	ctx context.Context,
	id uuid.UUID,
	name *string,
	isArchived *bool,
) error {
	m.UpdateCalls++

	return m.UpdateError
}

func (m *MockJarService) UpdateAllocations(
	ctx context.Context,
	updates []jar.JarAllocationUpdate,
) error {
	m.UpdateAllocationsCalls++
	m.UpdatedAllocations = updates

	return m.UpdateAllocationsError
}

func newTestHandler(service *MockJarService) *handler.JarHandler {
	logger := slog.New(
		slog.NewTextHandler(io.Discard, nil),
	)

	return handler.NewJarHandler(service, logger)
}

func withChiURLParam(
	req *http.Request,
	key string,
	value string,
) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)

	return req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
	)
}

func TestJarHandler_Create(t *testing.T) {
	createdJar := &jar.Jar{
		ID:              uuid.New(),
		Name:            "Needs",
		AllocationType:  jar.AllocationTypePercentage,
		AllocationValue: 50,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	tests := []struct {
		name        string
		body        string
		createError error
		createdJar  *jar.Jar
		wantStatus  int
		wantCode    string
		wantMessage string
		wantError   string
	}{
		{
			name:       "success",
			body:       `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createdJar: createdJar,
			wantStatus: http.StatusCreated,
		},
		{
			name:        "malformed json",
			body:        `{"name":"Needs"`,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST",
			wantMessage: "Request body contains invalid JSON",
			wantError:   "ERR_INVALID_JSON",
		},
		{
			name:        "unknown field",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50,"foo":"bar"}`,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_REQUEST",
			wantMessage: "Request body contains invalid JSON",
			wantError:   "ERR_INVALID_JSON",
		},
		{
			name:        "empty name",
			body:        `{"name":"","allocation_type":"percentage","allocation_value":50}`,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "Invalid jar request",
			wantError:   "ERR_INVALID_INPUT",
		},
		{
			name:        "whitespace name",
			body:        `{"name":"   ","allocation_type":"percentage","allocation_value":50}`,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "Invalid jar request",
			wantError:   "ERR_INVALID_INPUT",
		},
		{
			name:        "invalid allocation type",
			body:        `{"name":"Needs","allocation_type":"invalid","allocation_value":50}`,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "Invalid allocation type",
			wantError:   "ERR_INVALID_ALLOCATION_TYPE",
		},
		{
			name:        "duplicate name",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createError: jar.ErrJarNameExists,
			wantStatus:  http.StatusConflict,
			wantCode:    "CONFLICT",
			wantMessage: "Jar name already exists",
			wantError:   "ERR_JAR_EXISTS",
		},
		{
			name:        "invalid name from service",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createError: jar.ErrInvalidName,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "Invalid jar name",
			wantError:   "ERR_INVALID_NAME",
		},
		{
			name:        "invalid allocation type from service",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createError: jar.ErrInvalidAllocationType,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "Invalid allocation type",
			wantError:   "ERR_INVALID_ALLOCATION_TYPE",
		},
		{
			name:        "invalid allocation value",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":150}`,
			createError: jar.ErrInvalidAllocationVal,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: "Invalid allocation value",
			wantError:   "ERR_INVALID_ALLOCATION_VALUE",
		},
		{
			name:        "empty allocation configuration",
			body:        `{"name":"Needs","allocation_type":"remainder","allocation_value":0}`,
			createError: jar.ErrEmptyAllocationConfiguration,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: jar.ErrEmptyAllocationConfiguration.Error(),
			wantError:   "ERR_INVALID_ALLOCATION",
		},
		{
			name:        "no remainder",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createError: jar.ErrNoRemainder,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: jar.ErrNoRemainder.Error(),
			wantError:   "ERR_INVALID_ALLOCATION",
		},
		{
			name:        "multiple remainders",
			body:        `{"name":"Needs","allocation_type":"remainder","allocation_value":0}`,
			createError: jar.ErrMultipleRemainders,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: jar.ErrMultipleRemainders.Error(),
			wantError:   "ERR_INVALID_ALLOCATION",
		},
		{
			name:        "percentage exceeds limit",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createError: jar.ErrPercentageExceedsLimit,
			wantStatus:  http.StatusBadRequest,
			wantCode:    "VALIDATION_ERROR",
			wantMessage: jar.ErrPercentageExceedsLimit.Error(),
			wantError:   "ERR_INVALID_ALLOCATION",
		},
		{
			name:        "internal error",
			body:        `{"name":"Needs","allocation_type":"percentage","allocation_value":50}`,
			createError: errors.New("database error"),
			wantStatus:  http.StatusInternalServerError,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "Failed to create jar",
			wantError:   "ERR_INTERNAL_SERVER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockJarService{
				CreateError: tt.createError,
				CreatedJar:  tt.createdJar,
			}

			h := newTestHandler(service)

			req := httptest.NewRequest(
				http.MethodPost,
				"/jars",
				strings.NewReader(tt.body),
			)
			rec := httptest.NewRecorder()

			h.Create(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if tt.wantCode != "" {
				var got handler.ErrorResponse

				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				want := handler.ErrorResponse{
					Error:   tt.wantCode,
					Message: tt.wantMessage,
					Code:    tt.wantError,
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %#v, want %#v", got, want)
				}

				return
			}

			if service.CreateCalls != 1 {
				t.Fatalf(
					"expected 1 Create call, got %d",
					service.CreateCalls,
				)
			}

			var got handler.JarResponse

			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			want := handler.NewJarResponse(createdJar)

			if !reflect.DeepEqual(got, want) {
				t.Errorf(
					"unexpected response:\n got: %#v\nwant: %#v",
					got,
					want,
				)
			}
		})
	}
}

func TestJarHandler_List(t *testing.T) {
	j1 := &jar.Jar{
		ID:              uuid.New(),
		Name:            "Needs",
		AllocationType:  jar.AllocationTypePercentage,
		AllocationValue: 50,
	}

	j2 := &jar.Jar{
		ID:              uuid.New(),
		Name:            "Archived",
		AllocationType:  jar.AllocationTypeFixed,
		AllocationValue: 1000,
		IsArchived:      true,
	}

	tests := []struct {
		name        string
		jars        []*jar.Jar
		listErr     error
		wantStatus  int
		wantCode    string
		wantMessage string
		wantError   string
		wantLen     int
	}{
		{
			name:       "success",
			jars:       []*jar.Jar{j1, j2},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:        "service error",
			listErr:     errors.New("database error"),
			wantStatus:  http.StatusInternalServerError,
			wantCode:    "INTERNAL_ERROR",
			wantMessage: "Failed to list jars",
			wantError:   "ERR_INTERNAL_SERVER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockJarService{
				Jars:      tt.jars,
				ListError: tt.listErr,
			}

			h := newTestHandler(service)

			req := httptest.NewRequest(
				http.MethodGet,
				"/jars",
				nil,
			)
			rec := httptest.NewRecorder()

			h.List(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if service.ListCalls != 1 {
				t.Fatalf(
					"expected 1 List call, got %d",
					service.ListCalls,
				)
			}

			if tt.wantCode != "" {
				var got handler.ErrorResponse

				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				want := handler.ErrorResponse{
					Error:   tt.wantCode,
					Message: tt.wantMessage,
					Code:    tt.wantError,
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %#v, want %#v", got, want)
				}

				return
			}

			var got []handler.JarResponse

			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(got) != tt.wantLen {
				t.Fatalf(
					"expected %d jars, got %d",
					tt.wantLen,
					len(got),
				)
			}

			if got[0].ID != j1.ID {
				t.Errorf(
					"expected first jar ID %v, got %v",
					j1.ID,
					got[0].ID,
				)
			}

			if got[0].Name != j1.Name {
				t.Errorf(
					"expected first jar name %q, got %q",
					j1.Name,
					got[0].Name,
				)
			}

			if got[0].AllocationType != j1.AllocationType {
				t.Errorf(
					"expected first allocation type %q, got %q",
					j1.AllocationType,
					got[0].AllocationType,
				)
			}

			if got[0].AllocationValue != j1.AllocationValue {
				t.Errorf(
					"expected first allocation value %d, got %d",
					j1.AllocationValue,
					got[0].AllocationValue,
				)
			}

			if got[0].IsArchived {
				t.Error("expected first jar to be active")
			}

			if got[1].ID != j2.ID {
				t.Errorf(
					"expected second jar ID %v, got %v",
					j2.ID,
					got[1].ID,
				)
			}

			if !got[1].IsArchived {
				t.Error("expected second jar to be archived")
			}
		})
	}
}

func TestJarHandler_Patch(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name            string
		id              string
		body            string
		updateError     error
		wantStatus      int
		wantUpdateCalls int
		wantCode        string
		wantMessage     string
		wantError       string
	}{
		{
			name:            "update name",
			id:              id.String(),
			body:            `{"name":"Essentials"}`,
			wantStatus:      http.StatusOK,
			wantUpdateCalls: 1,
		},
		{
			name:            "archive jar",
			id:              id.String(),
			body:            `{"is_archived":true}`,
			wantStatus:      http.StatusOK,
			wantUpdateCalls: 1,
		},
		{
			name:            "update name and archive",
			id:              id.String(),
			body:            `{"name":"Essentials","is_archived":true}`,
			wantStatus:      http.StatusOK,
			wantUpdateCalls: 1,
		},
		{
			name:            "invalid id",
			id:              "invalid-id",
			body:            `{"name":"Essentials"}`,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 0,
			wantCode:        "INVALID_REQUEST",
			wantMessage:     "Invalid jar ID format",
			wantError:       "ERR_INVALID_ID",
		},
		{
			name:            "malformed json",
			id:              id.String(),
			body:            `{"name":"Essentials"`,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 0,
			wantCode:        "INVALID_REQUEST",
			wantMessage:     "Request body contains invalid JSON",
			wantError:       "ERR_INVALID_JSON",
		},
		{
			name:            "unknown field",
			id:              id.String(),
			body:            `{"name":"Essentials","foo":"bar"}`,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 0,
			wantCode:        "INVALID_REQUEST",
			wantMessage:     "Request body contains invalid JSON",
			wantError:       "ERR_INVALID_JSON",
		},
		{
			name:            "no fields",
			id:              id.String(),
			body:            `{}`,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 0,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     "At least one field must be provided",
			wantError:       "ERR_NO_FIELDS",
		},
		{
			name:            "empty name",
			id:              id.String(),
			body:            `{"name":""}`,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 0,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     "Invalid patch format inputs",
			wantError:       "ERR_INVALID_INPUT",
		},
		{
			name:            "whitespace name",
			id:              id.String(),
			body:            `{"name":"   "}`,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 0,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     "Invalid patch format inputs",
			wantError:       "ERR_INVALID_INPUT",
		},
		{
			name:            "jar not found",
			id:              id.String(),
			body:            `{"name":"Essentials"}`,
			updateError:     jar.ErrJarNotFound,
			wantStatus:      http.StatusNotFound,
			wantUpdateCalls: 1,
			wantCode:        "NOT_FOUND",
			wantMessage:     "Jar not found",
			wantError:       "ERR_JAR_NOT_FOUND",
		},
		{
			name:            "duplicate name",
			id:              id.String(),
			body:            `{"name":"Essentials"}`,
			updateError:     jar.ErrJarNameExists,
			wantStatus:      http.StatusConflict,
			wantUpdateCalls: 1,
			wantCode:        "CONFLICT",
			wantMessage:     "Jar name already exists",
			wantError:       "ERR_JAR_EXISTS",
		},
		{
			name:            "invalid name from service",
			id:              id.String(),
			body:            `{"name":"Essentials"}`,
			updateError:     jar.ErrInvalidName,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 1,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     "Invalid jar name",
			wantError:       "ERR_INVALID_NAME",
		},
		{
			name:            "empty allocation configuration",
			id:              id.String(),
			body:            `{"is_archived":true}`,
			updateError:     jar.ErrEmptyAllocationConfiguration,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 1,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     jar.ErrEmptyAllocationConfiguration.Error(),
			wantError:       "ERR_INVALID_ALLOCATION",
		},
		{
			name:            "no remainder",
			id:              id.String(),
			body:            `{"is_archived":true}`,
			updateError:     jar.ErrNoRemainder,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 1,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     jar.ErrNoRemainder.Error(),
			wantError:       "ERR_INVALID_ALLOCATION",
		},
		{
			name:            "multiple remainders",
			id:              id.String(),
			body:            `{"is_archived":true}`,
			updateError:     jar.ErrMultipleRemainders,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 1,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     jar.ErrMultipleRemainders.Error(),
			wantError:       "ERR_INVALID_ALLOCATION",
		},
		{
			name:            "percentage exceeds limit",
			id:              id.String(),
			body:            `{"name":"Essentials"}`,
			updateError:     jar.ErrPercentageExceedsLimit,
			wantStatus:      http.StatusBadRequest,
			wantUpdateCalls: 1,
			wantCode:        "VALIDATION_ERROR",
			wantMessage:     jar.ErrPercentageExceedsLimit.Error(),
			wantError:       "ERR_INVALID_ALLOCATION",
		},
		{
			name:            "internal error",
			id:              id.String(),
			body:            `{"name":"Essentials"}`,
			updateError:     errors.New("database error"),
			wantStatus:      http.StatusInternalServerError,
			wantUpdateCalls: 1,
			wantCode:        "INTERNAL_ERROR",
			wantMessage:     "Failed to update jar",
			wantError:       "ERR_INTERNAL_SERVER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockJarService{
				UpdateError: tt.updateError,
			}

			h := newTestHandler(service)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/jars/"+tt.id,
				strings.NewReader(tt.body),
			)

			req = withChiURLParam(req, "id", tt.id)

			rec := httptest.NewRecorder()

			h.Patch(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if service.UpdateCalls != tt.wantUpdateCalls {
				t.Fatalf(
					"expected %d Update calls, got %d",
					tt.wantUpdateCalls,
					service.UpdateCalls,
				)
			}

			if tt.wantCode != "" {
				var got handler.ErrorResponse

				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				want := handler.ErrorResponse{
					Error:   tt.wantCode,
					Message: tt.wantMessage,
					Code:    tt.wantError,
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %#v, want %#v", got, want)
				}

				return
			}

			var got handler.SuccessResponse

			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			want := handler.SuccessResponse{
				Message: "Jar updated successfully",
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}

func TestJarHandler_UpdateAllocations(t *testing.T) {
	validID1 := uuid.New()
	validID2 := uuid.New()

	validBody := `{
		"allocations": [
			{
				"id": "` + validID1.String() + `",
				"allocation_type": "percentage",
				"allocation_value": 50
			},
			{
				"id": "` + validID2.String() + `",
				"allocation_type": "fixed",
				"allocation_value": 500000
			}
		]
	}`

	tests := []struct {
		name                       string
		body                       string
		updateAllocationsError     error
		wantStatus                 int
		wantUpdateAllocationsCalls int
		wantCode                   string
		wantMessage                string
		wantError                  string
		wantAllocations            []jar.JarAllocationUpdate
	}{
		{
			name:                       "success",
			body:                       validBody,
			wantStatus:                 http.StatusOK,
			wantUpdateAllocationsCalls: 1,
			wantAllocations: []jar.JarAllocationUpdate{
				{
					ID:              validID1,
					AllocationType:  jar.AllocationTypePercentage,
					AllocationValue: 50,
				},
				{
					ID:              validID2,
					AllocationType:  jar.AllocationTypeFixed,
					AllocationValue: 500000,
				},
			},
		},
		{
			name:                       "malformed json",
			body:                       `{"allocations":[`,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 0,
			wantCode:                   "INVALID_REQUEST",
			wantMessage:                "Request body contains invalid JSON",
			wantError:                  "ERR_INVALID_JSON",
		},
		{
			name: "unknown field",
			body: `{
				"allocations": [],
				"foo": "bar"
			}`,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 0,
			wantCode:                   "INVALID_REQUEST",
			wantMessage:                "Request body contains invalid JSON",
			wantError:                  "ERR_INVALID_JSON",
		},
		{
			name:                       "missing allocations",
			body:                       `{}`,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 0,
			wantCode:                   "VALIDATION_ERROR",
			wantMessage:                "Invalid allocation update request",
			wantError:                  "ERR_INVALID_INPUT",
		},
		{
			name: "empty allocations",
			body: `{
				"allocations": []
			}`,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 0,
			wantCode:                   "VALIDATION_ERROR",
			wantMessage:                "Invalid allocation update request",
			wantError:                  "ERR_INVALID_INPUT",
		},
		{
			name: "invalid id",
			body: `{
				"allocations": [
					{
						"id": "invalid-id",
						"allocation_type": "percentage",
						"allocation_value": 50
					}
				]
			}`,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 0,
			wantCode:                   "INVALID_REQUEST",
			wantMessage:                "Request body contains invalid JSON",
			wantError:                  "ERR_INVALID_JSON",
		},
		{
			name: "invalid allocation type",
			body: `{
				"allocations": [
					{
						"id": "` + validID1.String() + `",
						"allocation_type": "invalid",
						"allocation_value": 50
					}
				]
			}`,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 0,
			wantCode:                   "VALIDATION_ERROR",
			wantMessage:                "Invalid allocation type",
			wantError:                  "ERR_INVALID_ALLOCATION_TYPE",
		},
		{
			name:                       "jar not found",
			body:                       validBody,
			updateAllocationsError:     jar.ErrJarNotFound,
			wantStatus:                 http.StatusNotFound,
			wantUpdateAllocationsCalls: 1,
			wantCode:                   "NOT_FOUND",
			wantMessage:                "Jar not found",
			wantError:                  "ERR_JAR_NOT_FOUND",
		},
		{
			name:                       "invalid allocation type from service",
			body:                       validBody,
			updateAllocationsError:     jar.ErrInvalidAllocationType,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 1,
			wantCode:                   "VALIDATION_ERROR",
			wantMessage:                "Invalid allocation type",
			wantError:                  "ERR_INVALID_ALLOCATION_TYPE",
		},
		{
			name:                       "invalid allocation value from service",
			body:                       validBody,
			updateAllocationsError:     jar.ErrInvalidAllocationVal,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 1,
			wantCode:                   "VALIDATION_ERROR",
			wantMessage:                "Invalid allocation value",
			wantError:                  "ERR_INVALID_ALLOCATION_VALUE",
		},
		{
			name:                       "invalid allocation configuration",
			body:                       validBody,
			updateAllocationsError:     jar.ErrNoRemainder,
			wantStatus:                 http.StatusBadRequest,
			wantUpdateAllocationsCalls: 1,
			wantCode:                   "VALIDATION_ERROR",
			wantMessage:                jar.ErrNoRemainder.Error(),
			wantError:                  "ERR_INVALID_ALLOCATION",
		},
		{
			name:                       "internal error",
			body:                       validBody,
			updateAllocationsError:     errors.New("database error"),
			wantStatus:                 http.StatusInternalServerError,
			wantUpdateAllocationsCalls: 1,
			wantCode:                   "INTERNAL_ERROR",
			wantMessage:                "Failed to update jar allocations",
			wantError:                  "ERR_INTERNAL_SERVER",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &MockJarService{
				UpdateAllocationsError: tt.updateAllocationsError,
			}

			h := newTestHandler(service)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/jars/allocations",
				strings.NewReader(tt.body),
			)

			rec := httptest.NewRecorder()

			h.UpdateAllocations(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if service.UpdateAllocationsCalls != tt.wantUpdateAllocationsCalls {
				t.Fatalf(
					"expected %d UpdateAllocations calls, got %d",
					tt.wantUpdateAllocationsCalls,
					service.UpdateAllocationsCalls,
				)
			}

			if tt.wantCode != "" {
				var got handler.ErrorResponse

				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				want := handler.ErrorResponse{
					Error:   tt.wantCode,
					Message: tt.wantMessage,
					Code:    tt.wantError,
				}

				if !reflect.DeepEqual(got, want) {
					t.Errorf("got %#v, want %#v", got, want)
				}

				return
			}

			if !reflect.DeepEqual(
				service.UpdatedAllocations,
				tt.wantAllocations,
			) {
				t.Errorf(
					"unexpected allocations:\n got: %#v\nwant: %#v",
					service.UpdatedAllocations,
					tt.wantAllocations,
				)
			}

			var got handler.SuccessResponse

			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			want := handler.SuccessResponse{
				Message: "Jar allocations updated successfully",
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}
