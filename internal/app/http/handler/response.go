package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/domain"
	"github.com/leenwood/market-auth-service/internal/core/usecase"
)

// UserResponse is returned by /register and /me.
type UserResponse struct {
	ID        uuid.UUID `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email"      example:"user@example.com"`
	Name      string    `json:"name"       example:"Ivan Ivanov"`
	Role      string    `json:"role"       example:"buyer"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:00:00Z"`
}

// TokenResponse is returned by /login and /refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"  example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4"`
	ExpiresIn    int64  `json:"expires_in"    example:"900"`
}

// GuestTokenResponse is returned by /guest.
type GuestTokenResponse struct {
	AccessToken string `json:"access_token"  example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	GuestUserID string `json:"guest_user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExpiresIn   int64  `json:"expires_in"    example:"604800"`
}

// ErrorResponse is the standard error body.
type ErrorResponse struct {
	Error string `json:"error" example:"email already registered"`
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func toTokenResponse(p *usecase.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
		ExpiresIn:    p.ExpiresIn,
	}
}
