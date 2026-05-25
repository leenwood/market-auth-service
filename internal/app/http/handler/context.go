package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/app/http/middleware"
)

func subjectFromCtx(c *gin.Context) uuid.UUID {
	return middleware.SubjectFromCtx(c)
}
