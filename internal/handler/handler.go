package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leenwood/market-auth-service/internal/service"
	"github.com/leenwood/market-auth-service/internal/token"
	goredis "github.com/redis/go-redis/v9"
)

type redisClient interface {
	Ping(ctx context.Context) *goredis.StatusCmd
}

type Handler struct {
	service      *service.AuthService
	tokenManager *token.Manager
	db           *pgxpool.Pool
	redis        redisClient
}

func New(
	svc *service.AuthService,
	tm *token.Manager,
	db *pgxpool.Pool,
	redis redisClient,
) *Handler {
	return &Handler{
		service:      svc,
		tokenManager: tm,
		db:           db,
		redis:        redis,
	}
}
