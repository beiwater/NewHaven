package httpapi

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/newhaven/backend-next/internal/apperr"
)

const (
	defaultRateLimit  = 60
	defaultRateWindow = time.Minute
)

// RateLimiter implements a per-IP sliding window rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]int64 // IP -> sorted request timestamps (UnixNano)
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a RateLimiter. If limit <= 0, defaultRateLimit is used.
// If window <= 0, defaultRateWindow is used.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = defaultRateLimit
	}
	if window <= 0 {
		window = defaultRateWindow
	}
	return &RateLimiter{
		requests: make(map[string][]int64),
		limit:    limit,
		window:   window,
	}
}

// Middleware returns an http.Handler that rate-limits each client IP to limit
// requests per window. Exceeding the limit produces a 429 response with
// apperr.KindRateLimited. Health endpoints (/healthz, /readyz) are always allowed.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always allow health and readiness checks without consuming a slot.
		switch r.URL.Path {
		case "/healthz", "/readyz":
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)
		now := time.Now().UnixNano()
		windowNanos := rl.window.Nanoseconds()
		cutoff := now - windowNanos

		rl.mu.Lock()
		times := rl.requests[ip]

		// Trim timestamps outside the current window.
		// Times are stored in chronological order.
		start := 0
		for start < len(times) && times[start] < cutoff {
			start++
		}
		times = times[start:]

		if len(times) >= rl.limit {
			rl.mu.Unlock()
			writeAppErr(w, apperr.RateLimited("too many requests"))
			return
		}

		times = append(times, now)
		rl.requests[ip] = times
		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the IP address from an http.Request, stripping the port.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Fallback: RemoteAddr may already be just an IP (e.g. after chi's RealIP).
		return r.RemoteAddr
	}
	return host
}
