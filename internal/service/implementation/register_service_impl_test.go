package implementation

import (
	"context"
	"testing"

	"auth/internal/dto"
	"auth/internal/model"
	"auth/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	s := NewRegisterService(mockRepo)

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Simulation : l'utilisateur n'existe pas
	mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(false, nil)
	// Simulation : création réussie
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := s.Execute(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Email, user.User.Email)
	mockRepo.AssertExpectations(t)
}

func TestRegisterService_Execute_UserAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepo)
	s := NewRegisterService(mockRepo)

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
