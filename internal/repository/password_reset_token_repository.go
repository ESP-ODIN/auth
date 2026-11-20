package repository

import (
	"auth/internal/model"
	"context"

	"github.com/google/uuid"
)

type PasswordResetTokenRepository interface {
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error)
	MarkUsedAndInvalidateOthers(ctx context.Context, tokenID, userID uuid.UUID) error
}
