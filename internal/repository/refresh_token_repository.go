package repository

import (
	"auth/internal/model"
	"context"
)

// RefreshTokenRepository ne manipule que des hashs : le token en clair n'atteint jamais la base.
type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *model.RefreshToken) error
	GetByToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error
}
