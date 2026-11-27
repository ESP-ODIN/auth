package service

import (
	"auth/internal/dto"
	"context"
)

type ForgotPasswordServiceInterface interface {
	Execute(ctx context.Context, req *dto.ForgotPasswordRequest) error
}
