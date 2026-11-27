package router

import (
	_ "auth/docs"
	"auth/internal/handler"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	registerHandler *handler.RegisterHandler,
	loginHandler *handler.LoginHandler,
	refreshTokenHandler *handler.RefreshTokenHandler,
	resetPasswordHandler *handler.ResetPasswordHandler,
	forgotPasswordHandler *handler.ForgotPasswordHandler,
	rateLimiter gin.HandlerFunc,
	jwksHandler gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/.well-known/jwks.json", jwksHandler)

	api := r.Group("/api/v1/auth")
	{
		api.POST("/register", rateLimiter, registerHandler.Handle)
		api.POST("/login", rateLimiter, loginHandler.Handle)
		api.POST("/refresh", rateLimiter, refreshTokenHandler.Handle)
		api.POST("/reset-password", rateLimiter, resetPasswordHandler.Handle)
		api.POST("/forgot-password", rateLimiter, forgotPasswordHandler.Handle)
	}

	return r
}
