package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/model"
	"auth/internal/testutil"
	"auth/internal/utils"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testRefreshTokenTTL = 24 * time.Hour

func TestLoginService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockRefreshRepo := new(MockRefreshTokenRepo)
	signer, publicKey := testutil.Signer(t)
	s := NewLoginService(mockRepo, mockRefreshRepo, signer, testRefreshTokenTTL)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Simulation : l'utilisateur existe avec un vrai hash bcrypt du mot de passe
	hashedPassword, err := utils.HashPassword(req.Password)
	require.NoError(t, err)
	storedUser := &model.User{ID: uuid.New(), Email: req.Email, Password: hashedPassword, IsActive: true}
	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(storedUser, nil)
	// On capture ce que le service enregistre en base pour le comparer à la réponse
	var stored *model.RefreshToken
	mockRefreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).
		Run(func(args mock.Arguments) { stored = args.Get(1).(*model.RefreshToken) }).
		Return(nil)

	before := time.Now()
	resp, err := s.Login(context.Background(), req)
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Refresh token : renvoyé en clair au client, seul son hash est enregistré
	require.NotEmpty(t, resp.RefreshToken)
	require.NotNil(t, stored)
	assert.Equal(t, utils.HashRefreshToken(resp.RefreshToken), stored.TokenHash)
	assert.NotEqual(t, resp.RefreshToken, stored.TokenHash)
	assert.Equal(t, storedUser.ID, stored.UserID)
	assert.NotEqual(t, uuid.Nil, stored.ID)
	assert.Nil(t, stored.RevokedAt)
	assert.False(t, stored.ExpiresAt.Before(before.Add(testRefreshTokenTTL)))
	assert.False(t, stored.ExpiresAt.After(after.Add(testRefreshTokenTTL)))
	mockRefreshRepo.AssertExpectations(t)
	assert.Equal(t, dto.NewUserResponse(storedUser), resp.User)
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(resp.AccessToken, claims, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"}))
	require.NoError(t, err)
	assert.Equal(t, storedUser.ID, claims.ID)
	assert.Equal(t, storedUser.Email, claims.Email)
	mockRepo.AssertExpectations(t)
}

func TestLoginService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockRefreshRepo := new(MockRefreshTokenRepo)
	signer, _ := testutil.Signer(t)
	s := NewLoginService(mockRepo, mockRefreshRepo, signer, testRefreshTokenTTL)

	req := &dto.LoginRequest{
		Email:    "unknown@example.com",
		Password: "password123",
	}

	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, appErrors.ErrorUserNotFound)

	user, err := s.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, appErrors.ErrorInvalidCredentials, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
	mockRefreshRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestLoginService_Login_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockRefreshRepo := new(MockRefreshTokenRepo)
	signer, _ := testutil.Signer(t)
	s := NewLoginService(mockRepo, mockRefreshRepo, signer, testRefreshTokenTTL)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	hashedPassword, err := utils.HashPassword("other-password")
	require.NoError(t, err)
	storedUser := &model.User{ID: uuid.New(), Email: req.Email, Password: hashedPassword, IsActive: true}
	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(storedUser, nil)

	resp, err := s.Login(context.Background(), req)

	assert.ErrorIs(t, err, appErrors.ErrorInvalidCredentials)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
	mockRefreshRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestLoginService_Login_RefreshTokenStorageFails(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockRefreshRepo := new(MockRefreshTokenRepo)
	signer, _ := testutil.Signer(t)
	s := NewLoginService(mockRepo, mockRefreshRepo, signer, testRefreshTokenTTL)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	require.NoError(t, err)
	storedUser := &model.User{ID: uuid.New(), Email: req.Email, Password: hashedPassword, IsActive: true}
	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(storedUser, nil)
	dbErr := errors.New("database unavailable")
	mockRefreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(dbErr)

	resp, err := s.Login(context.Background(), req)

	// Sans refresh token enregistré, le login échoue : aucun token n'est renvoyé
	assert.ErrorIs(t, err, dbErr)
	assert.Nil(t, resp)
	mockRefreshRepo.AssertExpectations(t)
}

func TestLoginService_Login_InactiveUser(t *testing.T) {
	mockRepo := new(MockUserRepo)
	mockRefreshRepo := new(MockRefreshTokenRepo)
	signer, _ := testutil.Signer(t)
	s := NewLoginService(mockRepo, mockRefreshRepo, signer, testRefreshTokenTTL)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	hashedPassword, err := utils.HashPassword(req.Password)
	require.NoError(t, err)
	// Bon mot de passe, mais compte désactivé
	storedUser := &model.User{ID: uuid.New(), Email: req.Email, Password: hashedPassword, IsActive: false}
	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(storedUser, nil)

	resp, err := s.Login(context.Background(), req)

	assert.ErrorIs(t, err, appErrors.ErrorInvalidCredentials)
	assert.Nil(t, resp)
	mockRefreshRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
