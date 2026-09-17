package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"

	"auth/internal/dto"
	"auth/internal/model"
	"auth/internal/repository"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email or username already exists")
)

type RegisterService struct {
	userRepo repository.UserRepository
}

func NewRegisterService(userRepo repository.UserRepository) *RegisterService {
	return &RegisterService{
		userRepo: userRepo,
	}
}

func (s *RegisterService) Execute(ctx context.Context, req dto.RegisterRequest) (*model.User, error) {
	// 1. Vérifier si l'utilisateur existe déjà
	exists, err := s.userRepo.ExistsByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// 2. Hachage du mot de passe avec bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Préparation de l'entité model
	newUser := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user", // Rôle par défaut
		IsActive:     true,
	}

	// 4. Insertion en BDD Neon
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
