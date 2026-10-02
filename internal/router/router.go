package router

import (
	"auth/internal/handler"
	_ "auth/docs"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

func SetupRouter(
	registerHandler *handler.RegisterHandler,
	rateLimiter gin.HandlerFunc,
	jwksHandler gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/.well-known/jwks.json", jwksHandler)

	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", rateLimiter, registerHandler.Handle)
	}

	return r
}
