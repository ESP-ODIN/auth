package service

import (
	"auth/internal/dto"
	"context"
	
)

type RegisterServiceInterface interface {
	Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
}

type LoginServiceInterface interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
}
