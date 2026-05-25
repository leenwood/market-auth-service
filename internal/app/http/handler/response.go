package handler

import (
	"time"

	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/domain"
	"github.com/leenwood/market-auth-service/internal/core/usecase"
)

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

func toTokenResponse(p *usecase.TokenPair) tokenResponse {
	return tokenResponse{
		AccessToken:  p.AccessToken,
		RefreshToken: p.RefreshToken,
		ExpiresIn:    p.ExpiresIn,
	}
}
