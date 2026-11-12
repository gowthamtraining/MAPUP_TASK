package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mapup-geofence/internal/middleware"
)

func TestRateLimiter(t *testing.T) {
	// 5 requests allowed burst, 1 refill/sec
	rl := middleware.NewRateLimiter(1, 3, time.Second)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.LimitMiddleware(nextHandler)

	// First 3 requests should pass (capacity is 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("POST", "/vehicles/location", nil)
		req.Header.Set("X-Vehicle-ID", "test-truck-1")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected request %d to succeed, got %d", i+1, rec.Code)
		}
		if rec.Header().Get("X-RateLimit-Limit") != "3" {
			t.Errorf("expected X-RateLimit-Limit 3, got %s", rec.Header().Get("X-RateLimit-Limit"))
		}
	}

	// 4th request should exceed rate limit
	req := httptest.NewRequest("POST", "/vehicles/location", nil)
	req.Header.Set("X-Vehicle-ID", "test-truck-1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") != "1" {
		t.Errorf("expected Retry-After header 1, got %s", rec.Header().Get("Retry-After"))
	}
}
