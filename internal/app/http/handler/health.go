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
	for _, p := range h.pingers {
		if err := p.Check.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": p.Name + " unavailable",
				"error":  err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *Handler) Metrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

func (h *Handler) JWKS(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", h.jwks.JWKS())
}
