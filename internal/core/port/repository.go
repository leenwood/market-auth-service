package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type TokenRepository interface {
	Save(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	Get(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
}
