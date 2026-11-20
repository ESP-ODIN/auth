package implementation

import (
	"auth/internal/dto"
	"auth/internal/errors"
	"auth/internal/repository"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordService struct {
	tokenRepo repository.PasswordResetTokenRepository
	userRepo  repository.UserRepository
	now       func() time.Time
}

func NewResetPasswordService(
	tokenRepo repository.PasswordResetTokenRepository,
	userRepo repository.UserRepository,
) *ResetPasswordService {
	return &ResetPasswordService{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		now:       time.Now,
	}
}

func (s *ResetPasswordService) Execute(ctx context.Context, req *dto.ResetPasswordRequest) error {
	hash := sha256Hash(req.Token)

	token, err := s.tokenRepo.FindByTokenHash(ctx, hash)
	if err != nil {
		return errors.ErrorInvalidOrExpiredToken
	}

	if token.Used || token.ExpiresAt.Before(s.now()) {
		return errors.ErrorInvalidOrExpiredToken
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, token.UserID, string(hashedPassword)); err != nil {
		return err
	}

	if err := s.tokenRepo.MarkUsedAndInvalidateOthers(ctx, token.ID, token.UserID); err != nil {
		return err
	}

	return nil
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
