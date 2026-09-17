package router

import (
	"auth/internal/handler"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	registerHandler *handler.RegisterHandler,
) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", registerHandler.Handle)
	}

	return r
}
