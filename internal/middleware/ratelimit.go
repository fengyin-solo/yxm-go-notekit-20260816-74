package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a sliding-window per-IP rate limiter.
// Limit is the max number of requests allowed per window.
type RateLimiter struct {
	limit  int
	window time.Duration
	mu     sync.Mutex
	buckets map[string]*slidingWindow
}

type slidingWindow struct {
	mu      sync.Mutex
	hits    []int64 // nanosecond timestamps of each hit
	window  time.Duration
}

// NewRateLimiter creates a rate limiter with the given limit per window duration.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 100
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*slidingWindow),
	}
}

// Limit returns middleware that enforces the rate limit by client IP.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Allow reports whether the given key may proceed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	sw, ok := rl.buckets[key]
	if !ok {
		sw = &slidingWindow{window: rl.window}
		rl.buckets[key] = sw
	}
	rl.mu.Unlock()

	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now().UnixNano()
	cutoff := now - rl.window.Nanoseconds()

	// Prune old hits
	n := 0
	for _, h := range sw.hits {
		if h > cutoff {
			break
		}
		n++
	}
	if n > 0 {
		sw.hits = sw.hits[n:]
	}

	if len(sw.hits) >= rl.limit {
		return false
	}
	sw.hits = append(sw.hits, now)
	return true
}

// LimitFromConfig creates a rate-limiter middleware from a requests-per-minute value.
// Pass 0 to disable.
func LimitFromConfig(rpm int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if rpm <= 0 {
			return next
		}
		rl := NewRateLimiter(rpm, time.Minute)
		return rl.Limit(next)
	}
}

func clientIP(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		for _, ip := range splitComma(fwd) {
			ip = trimSpace(ip)
			if ip != "" {
				return ip
			}
		}
	}
	if xip := r.Header.Get("X-Real-IP"); xip != "" {
		return trimSpace(xip)
	}
	return r.RemoteAddr
}

func splitComma(s string) []string {
	out := make([]string, 0, 2)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func trimSpace(s string) string {
	i, j := 0, len(s)-1
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j > i && (s[j] == ' ' || s[j] == '\t') {
		j--
	}
	return s[i : j+1]
}
