package service

import (
	"auth/internal/dto"
	"context"
)

type ResetPasswordServiceInterface interface {
	Execute(ctx context.Context, req *dto.ResetPasswordRequest) error
}
