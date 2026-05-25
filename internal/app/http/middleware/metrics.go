package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leenwood/market-auth-service/internal/platform/metrics"
)

func Prometheus(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		m.HTTPRequestsTotal.
			WithLabelValues(c.Request.Method, c.FullPath(), strconv.Itoa(c.Writer.Status())).
			Inc()
		m.HTTPRequestDuration.
			WithLabelValues(c.Request.Method, c.FullPath()).
			Observe(time.Since(start).Seconds())
	}
}
