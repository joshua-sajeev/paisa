// Package http provides HTTP adapters for the application.
package http

import (
	"log/slog"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// loginRateLimiter limits authentication attempts independently per client IP.
type loginRateLimiter struct {
	mu      sync.Mutex
	clients map[string]*rate.Limiter
	rate    rate.Limit
	burst   int
}

// newLoginRateLimiter creates a per-IP login rate limiter.
//
// ratePerMinute specifies the sustained number of requests allowed per minute.
// burst specifies the maximum number of requests allowed in a short burst.
func newLoginRateLimiter(ratePerMinute int, burst int) *loginRateLimiter {
	return &loginRateLimiter{
		clients: make(map[string]*rate.Limiter),
		rate:    rate.Limit(ratePerMinute) / 60,
		burst:   burst,
	}
}

// Allow reports whether the client identified by ip is allowed to make
// another login attempt.
func (l *loginRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, ok := l.clients[ip]
	if !ok {
		limiter = rate.NewLimiter(l.rate, l.burst)
		l.clients[ip] = limiter
	}

	return limiter.Allow()
}

// LoginRateLimitMiddleware limits requests to the login endpoint based on
// the client's IP address.
//
// Requests that exceed the configured limit receive HTTP 429 Too Many Requests.
func LoginRateLimitMiddleware(
	limiter *loginRateLimiter,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			if !limiter.Allow(ip) {
				logger.WarnContext(
					r.Context(),
					"login rate limit exceeded",
				)

				writeErrorJSON(
					w,
					http.StatusTooManyRequests,
					"TOO_MANY_REQUESTS",
					"Too many login attempts. Please try again later.",
					"ERR_RATE_LIMITED",
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// clientIP returns the client's IP address from the request's remote address.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
