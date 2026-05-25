package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Health godoc
// @Summary      Liveness probe
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary      Readiness probe — checks Postgres and Redis
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      503  {object}  ErrorResponse
// @Router       /ready [get]
func (h *Handler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	for _, p := range h.pingers {
		if err := p.Check.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error: p.Name + " unavailable: " + err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Metrics returns the Prometheus scrape handler.
func (h *Handler) Metrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// JWKS godoc
// @Summary      JWKS — RSA public key for JWT verification
// @Tags         system
// @Produce      json
// @Success      200  {object}  object
// @Router       /.well-known/jwks.json [get]
func (h *Handler) JWKS(c *gin.Context) {
	c.Data(http.StatusOK, "application/json", h.jwks.JWKS())
}
