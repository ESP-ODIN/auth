package implementation

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"auth/internal/dto"
	"auth/internal/errors"
	"auth/internal/model"
	"auth/internal/repository"
	"auth/internal/utils"

	"github.com/google/uuid"
)

// MinimumAge est l'âge minimum requis pour s'inscrire.
const MinimumAge = 18

type RegisterService struct {
	userRepo repository.UserRepository
	signer   *utils.JWTSigner
	// now renvoie la date actuelle. Remplaçable dans les tests pour
	// obtenir des résultats stables quel que soit le jour d'exécution.
	now func() time.Time
}

func NewRegisterService(userRepo repository.UserRepository, signer *utils.JWTSigner) *RegisterService {
	return &RegisterService{
		userRepo: userRepo,
		signer:   signer,
		now:      time.Now,
	}
}

func (s *RegisterService) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 1. Conversion de la date de naissance (format déjà validé par le DTO)
	dateOfBirth, err := time.Parse(dto.DateLayout, req.DateOfBirth)
	if err != nil {
		return nil, err
	}

	// 2. Vérifier que l'utilisateur est majeur (avant tout appel BDD)
	if !isAdult(dateOfBirth, s.now()) {
		return nil, errors.ErrorUserUnderage
	}

	// 3. Vérifier que l'email n'est pas déjà utilisé
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrorEmailAlreadyExists
	}

	// 4. Vérifier que le pseudo n'est pas déjà utilisé
	exists, err = s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrorUsernameAlreadyExists
	}

	// 5. Hachage du mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 6. Préparation de l'entité model
	newUser := &model.User{
		ID:          uuid.New(),
		Email:       req.Email,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: dateOfBirth,
		Password:    string(hashedPassword),
		RoleID:      1, // Assurez-vous que l'ID 1 existe en BDD pour le rôle 'user'
		IsActive:    true,
	}

	// 7. Insertion en BDD
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// 8. Génération du JWT
	token, err := s.signer.GenerateJWT(newUser.ID, newUser.Email)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		User:  dto.NewUserResponse(newUser),
		Token: token,
	}, nil
}

// isAdult indique si une personne née à dateOfBirth a au moins MinimumAge ans à la date now.
// On calcule la date limite (aujourd'hui moins 18 ans) : il faut être né ce jour-là ou avant.
// Les deux dates sont ramenées à minuit UTC pour ne comparer que les jours.
func isAdult(dateOfBirth, now time.Time) bool {
	y, m, d := now.Date()
	limit := time.Date(y-MinimumAge, m, d, 0, 0, 0, 0, time.UTC)
	return !dateOfBirth.After(limit)
}
