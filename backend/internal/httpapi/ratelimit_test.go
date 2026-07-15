package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/config"
)

func TestRateLimiter_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_OverLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "10.0.0.2:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// 6th request — should be rate-limited
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != ErrorRateLimited {
		t.Fatalf("expected error code %q, got %q", ErrorRateLimited, resp.Error.Code)
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	window := 50 * time.Millisecond
	rl := NewRateLimiter(3, window)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "10.0.0.3:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("setup request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Next request should be blocked
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.3:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 before window reset, got %d", w.Code)
	}

	// Wait for the window to pass
	time.Sleep(window + 10*time.Millisecond)

	// Now it should succeed again
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.3:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 after window reset, got %d", w.Code)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust limit for IP A
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.RemoteAddr = "10.0.0.4:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("IP A request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// IP A should now be blocked
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.4:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("IP A after limit: expected 429, got %d", w.Code)
	}

	// IP B should still be allowed
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.5:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("IP B: expected 200, got %d", w.Code)
	}
}

func TestRateLimiter_IPWithoutPort(t *testing.T) {
	// Some middleware (e.g. chi's RealIP) may set RemoteAddr without port.
	rl := NewRateLimiter(2, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.6"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Second request with same IP (different port string but same IP after stripping)
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.6:9999"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Third request should be blocked (limit is 2)
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.RemoteAddr = "10.0.0.6:9999"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestRateLimiter_ConcurrentSafety(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var blocked int

	// Fire 50 concurrent requests from the same IP
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.RemoteAddr = "10.0.0.8:12345"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				mu.Lock()
				blocked++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if blocked == 50 {
		t.Fatal("all 50 concurrent requests blocked — likely a locking issue")
	}
}

// TestRateLimiter_NewRouterIntegration verifies that the rate limiter is wired
// into NewRouter when RateLimitEnabled is true, health routes are unaffected,
// and API routes still work for single requests.
func TestRateLimiter_NewRouterIntegration(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey:    "test-secret",
		RateLimitEnabled: true,
	}
	mux := NewRouter(cfg, &RouterHandlers{})

	// Health endpoints should return 200 without rate limiting
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz: expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("readyz: expected 200, got %d", w.Code)
	}

	// An unregistered API route should return 404, not 429 (rate limiter present
	// but passes a single request through).
	req = httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code == http.StatusTooManyRequests {
		t.Fatalf("first request to /api got 429 unexpectedly")
	}
}
