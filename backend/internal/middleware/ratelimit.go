package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
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

// NewRateLimiter initializes a rate limiter. E.g., rate=25, capacity=40, interval=1s
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

func (rl *RateLimiter) allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists {
		v = &visitor{
			tokens:     rl.capacity - 1,
			lastRefill: now,
		}
		rl.visitors[key] = v
		return true, v.tokens
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
		return true, v.tokens
	}

	return false, 0
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

// LimitMiddleware returns HTTP middleware applying rate limiting with standard RFC rate limit headers
func (rl *RateLimiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Key by Vehicle ID header if provided, or client IP
		key := r.Header.Get("X-Vehicle-ID")
		if key == "" {
			key = r.Header.Get("X-Forwarded-For")
		}
		if key == "" {
			key = r.RemoteAddr
		}

		allowed, remaining := rl.allow(key)

		// Set standard RateLimit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.capacity))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", "1")

		if !allowed {
			w.Header().Set("Retry-After", "1")
			elapsed := time.Since(start).Nanoseconds()
			if elapsed <= 0 {
				elapsed = 150
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   fmt.Sprintf("rate limit exceeded: max %d requests per second allowed", rl.rate),
				"time_ns": strconv.FormatInt(elapsed, 10),
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
