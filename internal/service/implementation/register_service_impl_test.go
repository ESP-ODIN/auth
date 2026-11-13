package implementation

import (
	"context"
	"testing"
	"time"

	"auth/internal/dto"
	"auth/internal/errors"
	"auth/internal/testutil"
	"auth/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fixedNow est la date "aujourd'hui" utilisée par tous les tests,
// pour que les résultats ne dépendent pas du jour d'exécution.
var fixedNow = time.Date(2026, time.November, 6, 15, 0, 0, 0, time.UTC)

// newTestRegisterService crée un service avec un mock et une date figée.
func newTestRegisterService(t *testing.T) (*RegisterService, *MockUserRepo, any) {
	mockRepo := new(MockUserRepo)
	signer, publicKey := testutil.Signer(t)
	s := NewRegisterService(mockRepo, signer)
	s.now = func() time.Time { return fixedNow }
	return s, mockRepo, publicKey
}

// validRegisterRequest renvoie une requête valide (utilisateur majeur).
func validRegisterRequest() dto.RegisterRequest {
	return dto.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		Username:    "testuser",
		FirstName:   "Test",
		LastName:    "User",
		DateOfBirth: "1990-01-01",
	}
}

func TestRegisterService_Execute_Success(t *testing.T) {
	s, mockRepo, publicKey := newTestRegisterService(t)
	req := validRegisterRequest()

	// Simulation : email et pseudo libres, création réussie
	mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(false, nil)
	mockRepo.On("ExistsByUsername", mock.Anything, req.Username).Return(false, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)

	user, err := s.Execute(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, user)
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(user.Token, claims, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"}))
	require.NoError(t, err)
	assert.Equal(t, user.User.ID, claims.ID.String())
	assert.Equal(t, user.User.Email, claims.Email)
	assert.Equal(t, req.Email, user.User.Email)
	assert.Equal(t, req.Username, user.User.Username)
	assert.Equal(t, req.FirstName, user.User.FirstName)
	assert.Equal(t, req.LastName, user.User.LastName)
	assert.Equal(t, req.DateOfBirth, user.User.DateOfBirth)
	mockRepo.AssertExpectations(t)
}

func TestRegisterService_Execute_EmailAlreadyExists(t *testing.T) {
	s, mockRepo, _ := newTestRegisterService(t)
	req := validRegisterRequest()

	// Simulation : l'email est déjà pris (le pseudo n'est jamais vérifié)
	mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(true, nil)

	user, err := s.Execute(context.Background(), req)

	assert.Equal(t, errors.ErrorEmailAlreadyExists, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "ExistsByUsername", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRegisterService_Execute_UsernameAlreadyExists(t *testing.T) {
	s, mockRepo, _ := newTestRegisterService(t)
	req := validRegisterRequest()

	// Simulation : email libre, mais pseudo déjà pris
	mockRepo.On("ExistsByEmail", mock.Anything, req.Email).Return(false, nil)
	mockRepo.On("ExistsByUsername", mock.Anything, req.Username).Return(true, nil)

	user, err := s.Execute(context.Background(), req)

	assert.Equal(t, errors.ErrorUsernameAlreadyExists, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRegisterService_Execute_Underage(t *testing.T) {
	s, mockRepo, _ := newTestRegisterService(t)
	req := validRegisterRequest()
	// Aura 18 ans demain (7 novembre 2026) : encore mineur aujourd'hui
	req.DateOfBirth = "2008-11-07"

	user, err := s.Execute(context.Background(), req)

	assert.Equal(t, errors.ErrorUserUnderage, err)
	assert.Nil(t, user)
	// La règle d'âge est vérifiée avant tout appel à la base
	mockRepo.AssertNotCalled(t, "ExistsByEmail", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "ExistsByUsername", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestIsAdult(t *testing.T) {
	tests := []struct {
		name        string
		dateOfBirth string
		want        bool
	}{
		{"18 ans aujourd'hui", "2008-11-06", true},
		{"18 ans demain", "2008-11-07", false},
		{"18 ans depuis hier", "2008-11-05", true},
		{"même année mais anniversaire passé", "2008-01-01", true},
		{"même année mais anniversaire à venir", "2008-12-31", false},
		{"largement majeur", "1990-01-01", true},
		{"date dans le futur", "2030-01-01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dob, err := time.Parse("2006-01-02", tt.dateOfBirth)
			require.NoError(t, err)
			assert.Equal(t, tt.want, isAdult(dob, fixedNow))
		})
	}
}
