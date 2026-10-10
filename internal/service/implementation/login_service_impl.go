package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/repository"
	"auth/internal/utils"
	"context"
	"errors"
)

type LoginServiceImpl struct {
	userRepository repository.UserRepository
}

func NewLoginService(userRepository repository.UserRepository) *LoginServiceImpl {
	return &LoginServiceImpl{userRepository: userRepository}
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

	return &dto.LoginResponse{}, nil
}
