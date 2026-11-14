package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/model"
	"auth/internal/repository"
	"auth/internal/utils"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RefreshTokenServiceImpl struct {
	userRepository         repository.UserRepository
	refreshTokenRepository repository.RefreshTokenRepository
	signer                 *utils.JWTSigner
	refreshTokenTTL        time.Duration
}

func NewRefreshTokenService(userRepository repository.UserRepository, refreshTokenRepository repository.RefreshTokenRepository, signer *utils.JWTSigner, refreshTokenTTL time.Duration) *RefreshTokenServiceImpl {
	return &RefreshTokenServiceImpl{
		userRepository:         userRepository,
		refreshTokenRepository: refreshTokenRepository,
		signer:                 signer,
		refreshTokenTTL:        refreshTokenTTL,
	}
}

// newRefreshToken fabrique un refresh token pour un utilisateur.
// Il renvoie le token en clair (pour le client) et la ligne à enregistrer, qui ne contient que son hash.
func newRefreshToken(userID uuid.UUID, ttl time.Duration) (string, *model.RefreshToken, error) {
	token, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", nil, err
	}
	return token, &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: utils.HashRefreshToken(token),
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// Refresh échange un refresh token valide contre un nouvel access token et un nouveau refresh token (rotation).
// Un token inconnu, expiré ou révoqué donne toujours ErrorInvalidRefreshToken, sans préciser la raison.
// Les erreurs techniques (base indisponible, signature) remontent telles quelles.
func (s *RefreshTokenServiceImpl) Refresh(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	tokenHash := utils.HashRefreshToken(req.RefreshToken)

	// 1. Le token doit exister
	current, err := s.refreshTokenRepository.GetByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, appErrors.ErrorRefreshTokenNotFound) {
			return nil, appErrors.ErrorInvalidRefreshToken
		}
		return nil, err
	}
	userID := current.UserID.String()

	// 2. Réutilisation d'un token déjà révoqué : signe de vol, on révoque tous les tokens de l'utilisateur
	if current.RevokedAt != nil {
		if err := s.refreshTokenRepository.RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
			return nil, err
		}
		return nil, appErrors.ErrorInvalidRefreshToken
	}

	// 3. Le token ne doit pas être expiré
	if !time.Now().Before(current.ExpiresAt) {
		return nil, appErrors.ErrorInvalidRefreshToken
	}

	// 4. L'utilisateur doit toujours exister et être actif
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrorUserNotFound) {
			return nil, appErrors.ErrorInvalidRefreshToken
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, appErrors.ErrorInvalidRefreshToken
	}

	// 5. On prépare les nouveaux tokens avant de toucher à la base :
	//    si la génération échoue, l'ancien refresh token reste utilisable
	accessToken, err := s.signer.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}
	refreshToken, record, err := newRefreshToken(user.ID, s.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	// 6. Rotation : l'ancien token est révoqué. La révocation ne réussit que s'il était encore actif ;
	//    si une requête simultanée l'a révoqué entre-temps, c'est aussi une réutilisation.
	if err := s.refreshTokenRepository.RevokeRefreshToken(ctx, tokenHash); err != nil {
		if errors.Is(err, appErrors.ErrorRefreshTokenNotFound) {
			if err := s.refreshTokenRepository.RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
				return nil, err
			}
			return nil, appErrors.ErrorInvalidRefreshToken
		}
		return nil, err
	}

	// 7. Le nouveau refresh token est enregistré (son hash uniquement)
	if err := s.refreshTokenRepository.Create(ctx, record); err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
