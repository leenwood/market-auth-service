package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/domain"
	goredis "github.com/redis/go-redis/v9"
)

const keyPrefix = "refresh:"

type TokenRepository struct {
	client *goredis.Client
}

func NewTokenRepository(client *goredis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

func (r *TokenRepository) Save(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error {
	return r.client.Set(ctx, keyPrefix+token, userID.String(), ttl).Err()
}

func (r *TokenRepository) Get(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := r.client.Get(ctx, keyPrefix+token).Result()
	if errors.Is(err, goredis.Nil) {
		return uuid.Nil, domain.ErrTokenNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("redis get: %w", err)
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user_id in redis: %w", err)
	}
	return id, nil
}

func (r *TokenRepository) Delete(ctx context.Context, token string) error {
	n, err := r.client.Del(ctx, keyPrefix+token).Result()
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	if n == 0 {
		return domain.ErrTokenNotFound
	}
	return nil
}
