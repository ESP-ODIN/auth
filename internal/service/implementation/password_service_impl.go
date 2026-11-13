package implementation

import (
	appErrors "auth/internal/errors"
	"auth/internal/repository"
	"auth/internal/utils"
	"context"
	"errors"
)

type PasswordServiceImpl struct {
	userRepository repository.UserRepository
	signer         *utils.JWTSigner
}

func NewPasswordService(userRepository repository.UserRepository, signer *utils.JWTSigner) *PasswordServiceImpl {
	return &PasswordServiceImpl{userRepository: userRepository, signer: signer}
	}

func (s *PasswordServiceImpl) ChangePassword(ctx context.Context, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepository.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, appErrors.ErrorUserNotFound) {
			return nil, appErrors.ErrorInvalidCredentials
		}
		return nil, err
	}
	if !utils.ComparePassword(user.Password, req.Password) {
			return nil, appErrors.ErrorInvalidCredentials
		}
	new_password, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	user.Password = new_password
	err = s.userRepository.Update(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
	
