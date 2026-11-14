package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/repository"
	"auth/internal/utils"
	"context"
	"errors"
	"time"
)

type LoginServiceImpl struct {
	userRepository         repository.UserRepository
	refreshTokenRepository repository.RefreshTokenRepository
	signer                 *utils.JWTSigner
	refreshTokenTTL        time.Duration
}

func NewLoginService(userRepository repository.UserRepository, refreshTokenRepository repository.RefreshTokenRepository, signer *utils.JWTSigner, refreshTokenTTL time.Duration) *LoginServiceImpl {
	return &LoginServiceImpl{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		signer:                 signer,
		refreshTokenTTL:        refreshTokenTTL,
	}
}

func (s *LoginServiceImpl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, appErrors.ErrorUserNotFound) {
			return nil, appErrors.ErrorInvalidCredentials
		}
		return nil, err
	}

	if !utils.ComparePassword(user.Password, req.Password) {
		return nil, appErrors.ErrorInvalidCredentials
	}

	// Un compte désactivé ne peut pas se connecter. Même erreur générique, vérifiée après le mot de passe,
	// pour ne pas révéler l'état d'un compte à quelqu'un qui n'en connaît pas le mot de passe.
	if !user.IsActive {
		return nil, appErrors.ErrorInvalidCredentials
	}

	token, err := s.signer.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Refresh token : le client reçoit le token en clair, la base ne garde que son hash.
	refreshToken, record, err := newRefreshToken(user.ID, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokenRepository.Create(ctx, record); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User:         dto.NewUserResponse(user),
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}
