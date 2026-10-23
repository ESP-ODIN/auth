package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/model"
	"auth/internal/testutil"
	"auth/internal/utils"
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoginService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	signer, publicKey := testutil.Signer(t)
	s := NewLoginService(mockRepo, signer)

	req := &dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Simulation : l'utilisateur existe avec un vrai hash bcrypt du mot de passe
	hashedPassword, err := utils.HashPassword(req.Password)
	require.NoError(t, err)
	storedUser := &model.User{ID: uuid.New(), Email: req.Email, Password: hashedPassword}
	mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(storedUser, nil)

	resp, err := s.Login(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, storedUser, resp.User)
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(resp.Token, claims, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"}))
	require.NoError(t, err)
	assert.Equal(t, storedUser.ID, claims.ID)
	assert.Equal(t, storedUser.Email, claims.Email)
	mockRepo.AssertExpectations(t)
}

func TestLoginService_Login_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepo)
	signer, _ := testutil.Signer(t)
	s := NewLoginService(mockRepo, signer)

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
}
