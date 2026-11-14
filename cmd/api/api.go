package main

import (
	"auth/internal/handler"
	"auth/internal/middleware"
	repoImpl "auth/internal/repository/implementation"
	"auth/internal/router"
	svcImpl "auth/internal/service/implementation"
	"auth/internal/utils"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

func routes(db *pgxpool.Pool, signer *utils.JWTSigner, refreshTokenTTL time.Duration) http.Handler {
	userRepo := repoImpl.NewUserRepository(db)
	refreshTokenRepo := repoImpl.NewRefreshTokenRepository(db)
	registerService := svcImpl.NewRegisterService(userRepo, signer)
	registerHandler := handler.NewRegisterHandler(registerService)
	loginService := svcImpl.NewLoginService(userRepo, refreshTokenRepo, signer, refreshTokenTTL)
	loginHandler := handler.NewLoginHandler(loginService)
	refreshTokenService := svcImpl.NewRefreshTokenService(userRepo, refreshTokenRepo, signer, refreshTokenTTL)
	refreshTokenHandler := handler.NewRefreshTokenHandler(refreshTokenService)

	// Rate limiting: 1 request per second, burst of 5
	rateLimiter := middleware.NewRateLimiter(rate.Every(1*time.Second), 5)

	return router.SetupRouter(registerHandler, loginHandler, refreshTokenHandler, rateLimiter.Limit(), handler.NewJWKSHandler(signer))
}
