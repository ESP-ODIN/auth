package implementation

import (
	"auth/internal/model"
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockUserRepo est un mock partagé pour UserRepository
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

func (m *MockUserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockRefreshTokenRepo est un mock partagé pour RefreshTokenRepository
type MockRefreshTokenRepo struct {
	mock.Mock
}

func (m *MockRefreshTokenRepo) Create(ctx context.Context, refreshToken *model.RefreshToken) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockRefreshTokenRepo) GetByToken(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *MockRefreshTokenRepo) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, userID, hashedPassword)
	return args.Error(0)
}
