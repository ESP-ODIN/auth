package implementation

import (
	"context"
	"os"

	"golang.org/x/crypto/bcrypt"

	"auth/internal/dto"
	"auth/internal/model"
	"auth/internal/repository"
	"auth/internal/service"
	"auth/internal/utils"

	"github.com/google/uuid"
)

type RegisterService struct {
	userRepo repository.UserRepository
}

func NewRegisterService(userRepo repository.UserRepository) *RegisterService {
	return &RegisterService{
		userRepo: userRepo,
	}
}

func (s *RegisterService) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 1. Vérifier si l'utilisateur existe déjà
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, service.ErrUserAlreadyExists
	}

	// 2. Hachage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. Préparation de l'entité model
	newUser := &model.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: string(hashedPassword),
		RoleID:   1, // Assurez-vous que l'ID 1 existe en BDD pour le rôle 'user'
		IsActive: true,
	}

	// 4. Insertion en BDD
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// 5. Génération du JWT
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	token, err := utils.GenerateJWT(newUser.ID, newUser.Email, secretKey)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		User:  newUser,
		Token: token,
	}, nil
}
