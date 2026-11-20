package implementation

import (
	"auth/internal/errors"
	"auth/internal/model"
	"context"
	stdErrors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetTokenRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewPasswordResetTokenRepository(db *pgxpool.Pool) *PasswordResetTokenRepositoryImpl {
	return &PasswordResetTokenRepositoryImpl{db: db}
}

func (r *PasswordResetTokenRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (*model.PasswordResetToken, error) {
	var t model.PasswordResetToken
	query := `SELECT id, user_id, token_hash, expires_at, used, created_at
	          FROM public.password_reset_token WHERE token_hash = $1`
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.Used, &t.CreatedAt,
	)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.ErrorInvalidOrExpiredToken
		}
		return nil, err
	}
	return &t, nil
}

func (r *PasswordResetTokenRepositoryImpl) MarkUsedAndInvalidateOthers(ctx context.Context, tokenID, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE public.password_reset_token SET used = TRUE WHERE id = $1`, tokenID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE public.password_reset_token SET used = TRUE WHERE user_id = $1 AND id != $2 AND used = FALSE`,
		userID, tokenID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
