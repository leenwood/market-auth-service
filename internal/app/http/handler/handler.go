package handler

import (
	"context"

	"github.com/leenwood/market-auth-service/internal/core/usecase"
	"github.com/leenwood/market-auth-service/internal/platform/metrics"
)

// Pinger is satisfied by any dependency that exposes liveness.
type Pinger interface {
	Ping(ctx context.Context) error
}

// NamedPinger pairs a dependency name with its health check.
type NamedPinger struct {
	Name  string
	Check Pinger
}

type jwksProvider interface {
	JWKS() []byte
}

type Handler struct {
	auth    *usecase.AuthUseCase
	jwks    jwksProvider
	m       *metrics.Metrics
	pingers []NamedPinger
}

func New(
	auth *usecase.AuthUseCase,
	jwks jwksProvider,
	m *metrics.Metrics,
	pingers ...NamedPinger,
) *Handler {
	return &Handler{
		auth:    auth,
		jwks:    jwks,
		m:       m,
		pingers: pingers,
	}
}
