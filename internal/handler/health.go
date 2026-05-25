package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Ready(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unavailable", "error": err.Error()})
		return
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "redis unavailable", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) Metrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

func (h *Handler) JWKS(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", h.tokenManager.JWKS())
}
