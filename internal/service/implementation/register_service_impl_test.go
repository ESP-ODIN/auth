package implementation

import (
	"context"
	"testing"

	"auth/internal/dto"
	"auth/internal/model"
	"auth/internal/service"
	"auth/internal/testutil"
	"auth/internal/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock du repository
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestRegisterService_Execute_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)
	signer, publicKey := testutil.Signer(t)
	s := NewRegisterService(mockRepo, signer)

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Simulation : l'utilisateur n'existe pas
	mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(false, nil)
	// Simulation : création réussie
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := s.Execute(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, user)
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(user.Token, claims, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"}))
	require.NoError(t, err)
	assert.Equal(t, user.User.ID, claims.ID)
	assert.Equal(t, user.User.Email, claims.Email)
	assert.Equal(t, req.Email, user.User.Email)
	mockRepo.AssertExpectations(t)
}

func TestRegisterService_Execute_UserAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepo)
	signer, _ := testutil.Signer(t)
	s := NewRegisterService(mockRepo, signer)

	req := dto.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	// Simulation : l'utilisateur existe déjà
	mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(true, nil)

	user, err := s.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, service.ErrUserAlreadyExists, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}
