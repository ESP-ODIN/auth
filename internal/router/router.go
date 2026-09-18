package router

import (
	"auth/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	registerHandler *handler.RegisterHandler,
	rateLimiter gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", rateLimiter, registerHandler.Handle)
	}

	return r
}
