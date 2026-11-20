package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"
)

type ForgotPasswordService struct {
	userRepo  repository.UserRepository
	tokenRepo repository.PasswordResetTokenRepository
	email     service.EmailSender
}

func NewForgotPasswordService(
	userRepo repository.UserRepository,
	tokenRepo repository.PasswordResetTokenRepository,
	email service.EmailSender,
) *ForgotPasswordService {
	return &ForgotPasswordService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		email:     email,
	}
}

func (s *ForgotPasswordService) Execute(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	start := time.Now()
	defer func() {
		// Ensure comparable response time whether the email exists or not
		elapsed := time.Since(start)
		if elapsed < 200*time.Millisecond {
			time.Sleep(200*time.Millisecond - elapsed)
		}
	}()

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, appErrors.ErrorUserNotFound) {
			return nil
		}
		log.Printf("forgot-password: find user: %v", err)
		return nil
	}

	if !user.IsActive {
		return nil
	}

	rawToken, err := generateSecureToken(32)
	if err != nil {
		log.Printf("forgot-password: generate token: %v", err)
		return nil
	}

	hash := utils.SHA256Hash(rawToken)

	_, err = s.tokenRepo.Create(ctx, user.ID, hash)
	if err != nil {
		log.Printf("forgot-password: store token: %v", err)
		return nil
	}

	if err := s.email.SendResetEmail(user.Email, rawToken); err != nil {
		log.Printf("forgot-password: send email: %v", err)
	}

	return nil
}

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

