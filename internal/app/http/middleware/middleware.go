package middleware

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/leenwood/market-auth-service/internal/platform/logger"
	"github.com/leenwood/market-auth-service/internal/platform/metrics"
)

func Logger(log *slog.Logger, m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		span := trace.SpanFromContext(c.Request.Context())
		tid := span.SpanContext().TraceID().String()
		ctx := logger.WithTraceID(c.Request.Context(), tid)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		dur := time.Since(start)
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		logger.FromContext(ctx, log).Info("request",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", dur.Milliseconds(),
		)
		if m != nil {
			m.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
			m.HTTPRequestDuration.WithLabelValues(method, path).Observe(dur.Seconds())
		}
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		ctx := logger.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}
