package middleware

import (
	"net/http"
	"sync"
	"time"

	"mapup-geofence/internal/handler"
)

// RateLimiter implements an in-memory token bucket rate limiter per client/vehicle
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // tokens added per interval
	capacity int           // bucket max capacity
	interval time.Duration // refill interval
}

type visitor struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter initializes a rate limiter. E.g., rate=20, capacity=30, interval=1s
func NewRateLimiter(rate, capacity int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		capacity: capacity,
		interval: interval,
	}

	// Periodic cleanup of stale visitors
	go rl.cleanupRoutine(5 * time.Minute)

	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		rl.visitors[key] = &visitor{
			tokens:     rl.capacity - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on elapsed intervals
	elapsed := now.Sub(v.lastRefill)
	intervalsPassed := int(elapsed / rl.interval)
	if intervalsPassed > 0 {
		v.tokens += intervalsPassed * rl.rate
		if v.tokens > rl.capacity {
			v.tokens = rl.capacity
		}
		v.lastRefill = now
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastRefill) > 10*time.Minute {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// LimitMiddleware returns HTTP middleware applying rate limiting
func (rl *RateLimiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Key by RemoteAddr or client IP
		key := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			key = xff
		}

		if !rl.allow(key) {
			w.Header().Set("Retry-After", "1")
			handler.RespondError(w, http.StatusTooManyRequests, "rate limit exceeded: please reduce location ping frequency", start)
			return
		}

		next.ServeHTTP(w, r)
	})
}
