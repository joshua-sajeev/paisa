package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("Short response"))
	})

	loggedHandler := RequestLogger(logger)(nextHandler)

	req := httptest.NewRequest(http.MethodPost, "/test-path", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	ctx := context.WithValue(
		req.Context(),
		middleware.RequestIDKey,
		"test-req-id",
	)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	loggedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, "Short response", rec.Body.String())

	var logMap map[string]any
	err := json.Unmarshal(buf.Bytes(), &logMap)
	assert.NoError(t, err)

	assert.Equal(t, "HTTP Request", logMap["msg"])
	assert.Equal(t, "POST", logMap["method"])
	assert.Equal(t, "/test-path", logMap["path"])
	assert.Equal(t, "127.0.0.1:12345", logMap["remote_addr"])
	assert.Equal(t, float64(418), logMap["status"])
	assert.Equal(t, float64(14), logMap["bytes"])
	assert.Equal(t, "test-req-id", logMap["request_id"])
	assert.Contains(t, logMap, "duration")
}
