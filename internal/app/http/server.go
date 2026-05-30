package apphttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/leenwood/market-auth-service/docs/swagger" // register generated swagger spec
	"github.com/leenwood/market-auth-service/internal/app/http/handler"
	"github.com/leenwood/market-auth-service/internal/app/http/middleware"
	"github.com/leenwood/market-auth-service/internal/platform/metrics"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Debug        bool
}

func NewServer(cfg Config, h *handler.Handler, m *metrics.Metrics, log *slog.Logger, authMiddleware gin.HandlerFunc) *http.Server {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Logger(log, m))

	// system
	engine.GET("/health", h.Health)
	engine.GET("/ready", h.Ready)
	engine.GET("/metrics", h.Metrics())
	engine.GET("/.well-known/jwks.json", h.JWKS)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// auth API
	v1 := engine.Group("/api/v1/auth")
	{
		v1.POST("/register", h.Register)
		v1.POST("/login", h.Login)
		v1.POST("/refresh", h.Refresh)
		v1.POST("/logout", h.Logout)

		protected := v1.Group("/")
		protected.Use(authMiddleware)
		{
			protected.GET("/me", h.Me)
			protected.DELETE("/me", h.DeleteMe)
		}
	}

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
