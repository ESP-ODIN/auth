package repository

import (
	"auth/internal/model"
	"context"

	"github.com/google/uuid"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string) (*model.PasswordResetToken, error)
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
	MarkUsedAndInvalidateOthers(ctx context.Context, tokenID, userID uuid.UUID) error
}
