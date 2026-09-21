package main

import (
	"auth/internal/handler"
	"auth/internal/middleware"
	repoImpl "auth/internal/repository/implementation"
	"auth/internal/router"
	svcImpl "auth/internal/service/implementation"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

func routes(db *pgxpool.Pool) http.Handler {
	userRepo := repoImpl.NewUserRepository(db)
	registerService := svcImpl.NewRegisterService(userRepo)
	registerHandler := handler.NewRegisterHandler(registerService)

	// Rate limiting: 1 request per second, burst of 5
	rateLimiter := middleware.NewRateLimiter(rate.Every(1*time.Second), 5)

	return router.SetupRouter(registerHandler, rateLimiter.Limit())
}
