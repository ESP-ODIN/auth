package implementation

import (
	"auth/internal/errors"
	"auth/internal/model"
	"auth/internal/repository"
	"context"
	stdErrors "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Vérifié à la compilation : l'implémentation respecte bien l'interface.
var _ repository.RefreshTokenRepository = (*RefreshTokenRepositoryImpl)(nil)

type RefreshTokenRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepositoryImpl {
	return &RefreshTokenRepositoryImpl{db: db}
}

func (r *RefreshTokenRepositoryImpl) Create(ctx context.Context, refreshToken *model.RefreshToken) error {
	query := `
		INSERT INTO public.refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	// created_at est rempli par la base : on le relit pour que le modèle soit exact
	return r.db.QueryRow(ctx, query, refreshToken.ID, refreshToken.UserID, refreshToken.TokenHash, refreshToken.ExpiresAt).
		Scan(&refreshToken.CreatedAt)
}

func (r *RefreshTokenRepositoryImpl) GetByToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	query := "SELECT id, user_id, token_hash, expires_at, revoked_at, created_at FROM public.refresh_tokens WHERE token_hash = $1"
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.TokenHash, &refreshToken.ExpiresAt, &refreshToken.RevokedAt, &refreshToken.CreatedAt)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.ErrorRefreshTokenNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

// RevokeRefreshToken révoque un token encore actif.
// Si le token est inconnu ou déjà révoqué, aucune ligne n'est modifiée et ErrorRefreshTokenNotFound est renvoyée :
// deux requêtes simultanées avec le même token ne peuvent donc pas réussir toutes les deux.
func (r *RefreshTokenRepositoryImpl) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := "UPDATE public.refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL"
	cmdTag, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.ErrorRefreshTokenNotFound
	}
	return nil
}

// RevokeAllRefreshTokensForUser révoque tous les tokens encore actifs d'un utilisateur.
// N'en trouver aucun n'est pas une erreur.
func (r *RefreshTokenRepositoryImpl) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	query := "UPDATE public.refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL"
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
