package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				duration := time.Since(start)
				status := ww.Status()

				args := []any{
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", status),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Duration("duration", duration),
					slog.String("request_id", middleware.GetReqID(r.Context())),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
				}

				switch {
				case status >= 500:
					logger.ErrorContext(r.Context(), "HTTP Request", args...)
				case status >= 400:
					logger.WarnContext(r.Context(), "HTTP Request", args...)
				default:
					logger.InfoContext(r.Context(), "HTTP Request", args...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
