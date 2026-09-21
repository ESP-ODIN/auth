package service

import (
	"auth/internal/dto"
	"context"
	"errors"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type RegisterServiceInterface interface {
	Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
}
