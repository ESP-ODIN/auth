package implementation

import (
	"auth/internal/errors"
	"auth/internal/model"
	"context"
	stdErrors "errors"
	"time"

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

func (r *PasswordResetTokenRepositoryImpl) Create(ctx context.Context, userID uuid.UUID, tokenHash string) (*model.PasswordResetToken, error) {
	t := model.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
	}
	query := `INSERT INTO public.password_reset_token (id, user_id, token_hash, expires_at, used)
	          VALUES ($1, $2, $3, $4, $5) RETURNING created_at`
	err := r.db.QueryRow(ctx, query, t.ID, t.UserID, t.TokenHash, t.ExpiresAt, t.Used).Scan(&t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
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
