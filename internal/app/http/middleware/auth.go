package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/port"
)

const ctxKeyUserID = "userID"

// tokenParser is the minimal interface the middleware needs; satisfied by *infra/token.Manager.
type tokenParser interface {
	ParseAccessToken(string) (*port.ParsedToken, error)
}

func Auth(parser tokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		parsed, err := parser.ParseAccessToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		userID, err := uuid.Parse(parsed.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token subject"})
			return
		}

		c.Set(ctxKeyUserID, userID)
		c.Next()
	}
}

// SubjectFromCtx returns the authenticated user ID set by Auth middleware.
func SubjectFromCtx(c *gin.Context) uuid.UUID {
	return c.MustGet(ctxKeyUserID).(uuid.UUID)
}
