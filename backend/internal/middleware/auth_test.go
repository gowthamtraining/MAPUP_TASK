package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mapup-geofence/internal/middleware"
)

func TestAuthMiddleware_APIKey(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handlerToTest := middleware.AuthMiddleware(nextHandler)

	// Valid API key via header
	req := httptest.NewRequest("GET", "/geofences", nil)
	req.Header.Set("X-API-Key", middleware.DefaultAPIKey)
	rec := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK with valid API key, got %d", rec.Code)
	}

	// Invalid API key
	req = httptest.NewRequest("GET", "/geofences", nil)
	req.Header.Set("X-API-Key", "invalid-key-xyz")
	rec = httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with invalid API key, got %d", rec.Code)
	}
}

func TestAuthMiddleware_JWT(t *testing.T) {
	token, err := middleware.GenerateJWT("test-user", "admin", 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	claims, err := middleware.ValidateJWT(token)
	if err != nil {
		t.Fatalf("failed to validate JWT: %v", err)
	}
	if claims.Sub != "test-user" || claims.Role != "admin" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handlerToTest := middleware.AuthMiddleware(nextHandler)

	// Valid JWT Bearer token
	req := httptest.NewRequest("GET", "/vehicles", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid JWT, got %d", rec.Code)
	}

	// Expired JWT Token
	expiredToken, _ := middleware.GenerateJWT("test-user", "admin", -1*time.Hour)
	req = httptest.NewRequest("GET", "/vehicles", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec = httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with expired JWT, got %d", rec.Code)
	}
}

func TestAuthMiddleware_StrictEnforcement(t *testing.T) {
	os.Setenv("REQUIRE_AUTH", "true")
	defer os.Unsetenv("REQUIRE_AUTH")

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := middleware.AuthMiddleware(nextHandler)

	// Missing credentials under REQUIRE_AUTH=true should return 401
	req := httptest.NewRequest("GET", "/geofences", nil)
	rec := httptest.NewRecorder()
	handlerToTest.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when auth is required and no credentials, got %d", rec.Code)
	}
}
