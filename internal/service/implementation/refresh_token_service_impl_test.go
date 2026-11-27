package implementation

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"auth/internal/model"
	"auth/internal/testutil"
	"auth/internal/utils"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// refreshFixture prépare un service, ses mocks, un utilisateur actif et un refresh token valide.
type refreshFixture struct {
	service     *RefreshTokenServiceImpl
	userRepo    *MockUserRepo
	refreshRepo *MockRefreshTokenRepo
	user        *model.User
	token       string
	stored      *model.RefreshToken
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()
	signer, _ := testutil.Signer(t)
	userRepo := new(MockUserRepo)
	refreshRepo := new(MockRefreshTokenRepo)
	user := &model.User{ID: uuid.New(), Email: "test@example.com", IsActive: true}
	token, err := utils.GenerateRefreshToken()
	require.NoError(t, err)
	return &refreshFixture{
		service:     NewRefreshTokenService(userRepo, refreshRepo, signer, testRefreshTokenTTL),
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		user:        user,
		token:       token,
		stored: &model.RefreshToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			TokenHash: utils.HashRefreshToken(token),
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
}

func (f *refreshFixture) refresh() (*dto.RefreshTokenResponse, error) {
	return f.service.Refresh(context.Background(), &dto.RefreshTokenRequest{RefreshToken: f.token})
}

// assertNothingIssued vérifie qu'aucun nouveau refresh token n'a été enregistré.
func (f *refreshFixture) assertNothingIssued(t *testing.T) {
	t.Helper()
	f.refreshRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestRefreshTokenService_Refresh_Success(t *testing.T) {
	signer, publicKey := testutil.Signer(t)
	f := newRefreshFixture(t)
	f.service.signer = signer

	f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
	f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(f.user, nil)
	f.refreshRepo.On("RevokeRefreshToken", mock.Anything, f.stored.TokenHash).Return(nil)
	var created *model.RefreshToken
	f.refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).
		Run(func(args mock.Arguments) { created = args.Get(1).(*model.RefreshToken) }).
		Return(nil)

	before := time.Now()
	resp, err := f.refresh()
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Nouvel access token : un JWT valide pour le bon utilisateur
	claims := &utils.Claims{}
	_, err = jwt.ParseWithClaims(resp.AccessToken, claims, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"}))
	require.NoError(t, err)
	assert.Equal(t, f.user.ID, claims.ID)
	assert.Equal(t, f.user.Email, claims.Email)

	// Rotation : un nouveau refresh token, différent de l'ancien, dont seul le hash est enregistré
	require.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, f.token, resp.RefreshToken)
	require.NotNil(t, created)
	assert.Equal(t, utils.HashRefreshToken(resp.RefreshToken), created.TokenHash)
	assert.NotEqual(t, f.stored.TokenHash, created.TokenHash)
	assert.NotEqual(t, f.stored.ID, created.ID)
	assert.Equal(t, f.user.ID, created.UserID)
	assert.Nil(t, created.RevokedAt)
	assert.False(t, created.ExpiresAt.Before(before.Add(testRefreshTokenTTL)))
	assert.False(t, created.ExpiresAt.After(after.Add(testRefreshTokenTTL)))

	// L'ancien refresh token a été révoqué, et lui seul
	f.refreshRepo.AssertCalled(t, "RevokeRefreshToken", mock.Anything, f.stored.TokenHash)
	f.refreshRepo.AssertNotCalled(t, "RevokeAllRefreshTokensForUser", mock.Anything, mock.Anything)
	f.refreshRepo.AssertExpectations(t)
	f.userRepo.AssertExpectations(t)
}

func TestRefreshTokenService_Refresh_UnknownToken(t *testing.T) {
	f := newRefreshFixture(t)
	f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(nil, appErrors.ErrorRefreshTokenNotFound)

	resp, err := f.refresh()

	assert.ErrorIs(t, err, appErrors.ErrorInvalidRefreshToken)
	assert.Nil(t, resp)
	f.userRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	f.refreshRepo.AssertNotCalled(t, "RevokeRefreshToken", mock.Anything, mock.Anything)
	f.refreshRepo.AssertNotCalled(t, "RevokeAllRefreshTokensForUser", mock.Anything, mock.Anything)
	f.assertNothingIssued(t)
}

func TestRefreshTokenService_Refresh_ExpiredToken(t *testing.T) {
	f := newRefreshFixture(t)
	f.stored.ExpiresAt = time.Now().Add(-time.Second)
	f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)

	resp, err := f.refresh()

	assert.ErrorIs(t, err, appErrors.ErrorInvalidRefreshToken)
	assert.Nil(t, resp)
	f.refreshRepo.AssertNotCalled(t, "RevokeRefreshToken", mock.Anything, mock.Anything)
	f.assertNothingIssued(t)
}

// Un token déjà révoqué qui revient : toute la session de l'utilisateur est révoquée (détection de vol).
func TestRefreshTokenService_Refresh_RevokedTokenReuse(t *testing.T) {
	f := newRefreshFixture(t)
	revokedAt := time.Now().Add(-time.Minute)
	f.stored.RevokedAt = &revokedAt
	f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
	f.refreshRepo.On("RevokeAllRefreshTokensForUser", mock.Anything, f.user.ID.String()).Return(nil)

	resp, err := f.refresh()

	assert.ErrorIs(t, err, appErrors.ErrorInvalidRefreshToken)
	assert.Nil(t, resp)
	f.refreshRepo.AssertCalled(t, "RevokeAllRefreshTokensForUser", mock.Anything, f.user.ID.String())
	f.userRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
	f.assertNothingIssued(t)
}

// Deux requêtes simultanées avec le même token : celle qui perd la révocation est traitée comme une réutilisation.
func TestRefreshTokenService_Refresh_ConcurrentReuse(t *testing.T) {
	f := newRefreshFixture(t)
	f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
	f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(f.user, nil)
	f.refreshRepo.On("RevokeRefreshToken", mock.Anything, f.stored.TokenHash).Return(appErrors.ErrorRefreshTokenNotFound)
	f.refreshRepo.On("RevokeAllRefreshTokensForUser", mock.Anything, f.user.ID.String()).Return(nil)

	resp, err := f.refresh()

	assert.ErrorIs(t, err, appErrors.ErrorInvalidRefreshToken)
	assert.Nil(t, resp)
	f.refreshRepo.AssertCalled(t, "RevokeAllRefreshTokensForUser", mock.Anything, f.user.ID.String())
	f.assertNothingIssued(t)
}

func TestRefreshTokenService_Refresh_UserNoLongerAllowed(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(f *refreshFixture)
	}{
		{"utilisateur supprimé", func(f *refreshFixture) {
			f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(nil, appErrors.ErrorUserNotFound)
		}},
		{"utilisateur inactif", func(f *refreshFixture) {
			f.user.IsActive = false
			f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(f.user, nil)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newRefreshFixture(t)
			f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
			tc.setup(f)

			resp, err := f.refresh()

			assert.ErrorIs(t, err, appErrors.ErrorInvalidRefreshToken)
			assert.Nil(t, resp)
			f.refreshRepo.AssertNotCalled(t, "RevokeRefreshToken", mock.Anything, mock.Anything)
			f.assertNothingIssued(t)
		})
	}
}

// Une panne technique n'est pas un token invalide : l'erreur d'origine remonte (500 côté HTTP, pas 401).
func TestRefreshTokenService_Refresh_TechnicalErrors(t *testing.T) {
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name  string
		setup func(f *refreshFixture)
	}{
		{"recherche du token", func(f *refreshFixture) {
			f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(nil, dbErr)
		}},
		{"recherche de l'utilisateur", func(f *refreshFixture) {
			f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
			f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(nil, dbErr)
		}},
		{"révocation de l'ancien token", func(f *refreshFixture) {
			f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
			f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(f.user, nil)
			f.refreshRepo.On("RevokeRefreshToken", mock.Anything, f.stored.TokenHash).Return(dbErr)
		}},
		{"enregistrement du nouveau token", func(f *refreshFixture) {
			f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
			f.userRepo.On("FindByID", mock.Anything, f.user.ID.String()).Return(f.user, nil)
			f.refreshRepo.On("RevokeRefreshToken", mock.Anything, f.stored.TokenHash).Return(nil)
			f.refreshRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.RefreshToken")).Return(dbErr)
		}},
		{"révocation de la session après réutilisation", func(f *refreshFixture) {
			revokedAt := time.Now()
			f.stored.RevokedAt = &revokedAt
			f.refreshRepo.On("GetByToken", mock.Anything, f.stored.TokenHash).Return(f.stored, nil)
			f.refreshRepo.On("RevokeAllRefreshTokensForUser", mock.Anything, f.user.ID.String()).Return(dbErr)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newRefreshFixture(t)
			tc.setup(f)

			resp, err := f.refresh()

			assert.ErrorIs(t, err, dbErr)
			assert.NotErrorIs(t, err, appErrors.ErrorInvalidRefreshToken)
			assert.Nil(t, resp)
		})
	}
}
