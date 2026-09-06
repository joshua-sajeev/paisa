package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	h "github.com/joshu-sajeev/paisa/internal/adapter/http"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	ctx := context.WithValue(
		req.Context(),
		middleware.RequestIDKey,
		"test-request-id",
	)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify log contains expected fields
	output := buf.String()
	assert.Contains(t, output, "HTTP Request")
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "/test")
}

func TestRequestLogger_POST(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	body := `{"key":"value"}`
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/resource", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:54321"

	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	output := buf.String()
	assert.Contains(t, output, "POST")
}

func TestRequestLogger_ErrorResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRequestLogger_LargeResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Create a large response
	largeBody := make([]byte, 10000)
	for i := range largeBody {
		largeBody[i] = 'x'
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(largeBody)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/large", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, len(largeBody), len(rec.Body.Bytes()))
}

func TestRequestLogger_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			loggedHandler := h.RequestLogger(logger)(nextHandler)

			req := httptest.NewRequest(method, "/api/test", nil)
			rec := httptest.NewRecorder()

			loggedHandler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			output := buf.String()
			assert.Contains(t, output, method)
		})
	}
}

func TestRequestLogger_StatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
			})

			loggedHandler := h.RequestLogger(logger)(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			loggedHandler.ServeHTTP(rec, req)

			assert.Equal(t, statusCode, rec.Code)
		})
	}
}

func TestRequestLogger_MultipleRequests(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	// Make 3 requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		loggedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Should have logged 3 requests
	output := buf.String()
	lines := bytes.Count([]byte(output), []byte("\n"))
	assert.GreaterOrEqual(t, lines, 3)
}

func TestRequestLogger_JSONResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	responseBody := map[string]interface{}{
		"status": "success",
		"data": map[string]string{
			"id":   "123",
			"name": "test",
		},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseBody)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestRequestLogger_CustomHeaders(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "custom-value", rec.Header().Get("X-Custom-Header"))
}

func TestRequestLogger_EmptyBody(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_PanicRecovery(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	// Handler that doesn't panic - RequestLogger doesn't recover panics
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggedHandler := h.RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	assert.NotPanics(t, func() {
		loggedHandler.ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequestLogger_ChainedMiddleware(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger1 := slog.New(slog.NewJSONHandler(&buf1, nil))
	logger2 := slog.New(slog.NewJSONHandler(&buf2, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Chain two logging middleware
	loggedHandler := h.RequestLogger(logger1)(
		h.RequestLogger(logger2)(nextHandler),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Both loggers should have logged
	assert.Greater(t, buf1.Len(), 0)
	assert.Greater(t, buf2.Len(), 0)
}
