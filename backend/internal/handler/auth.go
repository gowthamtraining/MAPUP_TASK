package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"mapup-geofence/internal/middleware"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

type TokenRequest struct {
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
	ExpiresIn int    `json:"expires_in"`
	User      string `json:"user"`
	Role      string `json:"role"`
}

// GenerateToken handles POST /auth/token and POST /auth/login
func (h *AuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req TokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	user := "admin"
	if req.Username != "" {
		user = req.Username
	} else if req.ClientID != "" {
		user = req.ClientID
	}

	role := "admin"
	duration := 24 * time.Hour

	token, err := middleware.GenerateJWT(user, role, duration)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to generate token: "+err.Error(), start)
		return
	}

	res := TokenResponse{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: int(duration.Seconds()),
		User:      user,
		Role:      role,
	}

	RespondJSON(w, http.StatusOK, res, start)
}

// VerifyToken handles GET /auth/verify
func (h *AuthHandler) VerifyToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	user, _ := r.Context().Value(middleware.UserContextKey).(string)
	role, _ := r.Context().Value(middleware.RoleContextKey).(string)

	res := map[string]interface{}{
		"authenticated": true,
		"user":          user,
		"role":          role,
	}

	RespondJSON(w, http.StatusOK, res, start)
}
