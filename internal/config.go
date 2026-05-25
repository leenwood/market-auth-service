package internal

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseDSN      string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	JWTPrivateKey    string
	JWTPublicKey     string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	HTTPAddr         string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration
	LogLevel         string
	LogFormat        string
	OTELEnabled      bool
	OTELExporter     string
	OTELEndpoint     string
	OTELServiceName  string
}

func Load() (*Config, error) {
	cfg := &Config{
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8081"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		LogFormat:       getEnv("LOG_FORMAT", "json"),
		OTELExporter:    getEnv("OTEL_EXPORTER", "stdout"),
		OTELEndpoint:    getEnv("OTEL_ENDPOINT", ""),
		OTELServiceName: getEnv("OTEL_SERVICE_NAME", "auth-service"),
	}

	var err error

	cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("DATABASE_DSN is required")
	}

	cfg.JWTPrivateKey = os.Getenv("JWT_PRIVATE_KEY")
	if cfg.JWTPrivateKey == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY is required")
	}

	cfg.JWTPublicKey = os.Getenv("JWT_PUBLIC_KEY")
	if cfg.JWTPublicKey == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY is required")
	}

	cfg.RedisDB, err = strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_DB must be integer: %w", err)
	}

	if cfg.AccessTokenTTL, err = time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")); err != nil {
		return nil, fmt.Errorf("ACCESS_TOKEN_TTL invalid: %w", err)
	}
	if cfg.RefreshTokenTTL, err = time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "720h")); err != nil {
		return nil, fmt.Errorf("REFRESH_TOKEN_TTL invalid: %w", err)
	}
	if cfg.HTTPReadTimeout, err = time.ParseDuration(getEnv("HTTP_READ_TIMEOUT", "15s")); err != nil {
		return nil, fmt.Errorf("HTTP_READ_TIMEOUT invalid: %w", err)
	}
	if cfg.HTTPWriteTimeout, err = time.ParseDuration(getEnv("HTTP_WRITE_TIMEOUT", "15s")); err != nil {
		return nil, fmt.Errorf("HTTP_WRITE_TIMEOUT invalid: %w", err)
	}
	if cfg.HTTPIdleTimeout, err = time.ParseDuration(getEnv("HTTP_IDLE_TIMEOUT", "60s")); err != nil {
		return nil, fmt.Errorf("HTTP_IDLE_TIMEOUT invalid: %w", err)
	}

	cfg.OTELEnabled = getEnv("OTEL_ENABLED", "false") == "true"

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
