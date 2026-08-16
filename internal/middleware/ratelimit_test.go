package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)
	for i := 0; i < 3; i++ {
		if !rl.Allow("test-ip") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
	if rl.Allow("test-ip") {
		t.Error("4th request should be rejected")
	}
}

func TestRateLimiterSeparateIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)
	rl.Allow("ip-a")
	rl.Allow("ip-a")
	if !rl.Allow("ip-b") {
		t.Error("ip-b should have its own quota")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Hour)
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := rl.Limit(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want 200", i+1, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("request 3: status = %d, want 429", w.Code)
	}
}

func TestLimitFromConfigDisabled(t *testing.T) {
	h := LimitFromConfig(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Code != http.StatusOK {
		t.Errorf("disabled limiter should pass through, got %d", w.Code)
	}
}
