package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	internal "github.com/leenwood/market-auth-service/internal"
	apphttp "github.com/leenwood/market-auth-service/internal/app/http"
	"github.com/leenwood/market-auth-service/internal/app/http/handler"
	"github.com/leenwood/market-auth-service/internal/app/http/middleware"
	"github.com/leenwood/market-auth-service/internal/core/usecase"
	infratoken "github.com/leenwood/market-auth-service/internal/infra/token"
	infrapostgres "github.com/leenwood/market-auth-service/internal/infra/storage/postgres"
	infraredis "github.com/leenwood/market-auth-service/internal/infra/storage/redis"
	"github.com/leenwood/market-auth-service/internal/platform/logger"
	"github.com/leenwood/market-auth-service/internal/platform/metrics"
	"github.com/leenwood/market-auth-service/internal/platform/tracing"
	goredis "github.com/redis/go-redis/v9"
)

// pgPinger adapts pgxpool.Pool to handler.Pinger.
type pgPinger struct{ pool *pgxpool.Pool }

func (p *pgPinger) Ping(ctx context.Context) error { return p.pool.Ping(ctx) }

// redisPinger adapts goredis.Client to handler.Pinger.
type redisPinger struct{ client *goredis.Client }

func (r *redisPinger) Ping(ctx context.Context) error { return r.client.Ping(ctx).Err() }

func RunServer(ctx context.Context) error {
	cfg, err := internal.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	shutdownTracing, err := tracing.Init(ctx, tracing.Config{
		Enabled:     cfg.OTELEnabled,
		Exporter:    cfg.OTELExporter,
		Endpoint:    cfg.OTELEndpoint,
		ServiceName: cfg.OTELServiceName,
	})
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	slog.Info("postgres connected")

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping redis: %w", err)
	}
	slog.Info("redis connected")

	tokenManager, err := infratoken.NewManager(cfg.JWTPrivateKey, cfg.JWTPublicKey, cfg.AccessTokenTTL, cfg.GuestTokenTTL)
	if err != nil {
		return fmt.Errorf("init token manager: %w", err)
	}

	userRepo := infrapostgres.NewUserRepository(db)
	tokenRepo := infraredis.NewTokenRepository(redisClient)
	authUC := usecase.NewAuthUseCase(userRepo, tokenRepo, tokenManager, cfg.RefreshTokenTTL)

	m := metrics.New()

	h := handler.New(
		authUC,
		tokenManager,
		m,
		handler.NamedPinger{Name: "postgres", Check: &pgPinger{pool: db}},
		handler.NamedPinger{Name: "redis", Check: &redisPinger{client: redisClient}},
	)

	srv := apphttp.NewServer(
		apphttp.Config{
			Addr:         cfg.HTTPAddr,
			ReadTimeout:  cfg.HTTPReadTimeout,
			WriteTimeout: cfg.HTTPWriteTimeout,
			IdleTimeout:  cfg.HTTPIdleTimeout,
			Debug:        cfg.LogLevel == "debug",
		},
		h,
		m,
		log,
		middleware.Auth(tokenManager),
	)

	return runWithGracefulShutdown(ctx, srv, log, shutdownTracing)
}

func runWithGracefulShutdown(ctx context.Context, srv *http.Server, log *slog.Logger, shutdownTracing tracing.ShutdownFunc) error {
	serverErr := make(chan error, 1)
	go func() {
		log.Info("server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	_ = shutdownTracing(shutdownCtx)

	log.Info("server stopped")
	return nil
}

