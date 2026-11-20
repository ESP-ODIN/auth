package main

import (
	"auth/internal/handler"
	"auth/internal/middleware"
	repoImpl "auth/internal/repository/implementation"
	"auth/internal/router"
	"auth/internal/service"
	svcImpl "auth/internal/service/implementation"
	"auth/internal/utils"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

func routes(db *pgxpool.Pool, signer *utils.JWTSigner, emailSender service.EmailSender) http.Handler {
	userRepo := repoImpl.NewUserRepository(db)
	registerService := svcImpl.NewRegisterService(userRepo, signer)
	registerHandler := handler.NewRegisterHandler(registerService)
	loginService := svcImpl.NewLoginService(userRepo, signer)
	loginHandler := handler.NewLoginHandler(loginService)

	tokenRepo := repoImpl.NewPasswordResetTokenRepository(db)
	resetPasswordService := svcImpl.NewResetPasswordService(tokenRepo, userRepo)
	resetPasswordHandler := handler.NewResetPasswordHandler(resetPasswordService)

	forgotPasswordService := svcImpl.NewForgotPasswordService(userRepo, tokenRepo, emailSender)
	forgotPasswordHandler := handler.NewForgotPasswordHandler(forgotPasswordService)

	// Rate limiting: 1 request per second, burst of 5
	rateLimiter := middleware.NewRateLimiter(rate.Every(1*time.Second), 5)

	return router.SetupRouter(registerHandler, loginHandler, resetPasswordHandler, forgotPasswordHandler, rateLimiter.Limit(), handler.NewJWKSHandler(signer))
}
