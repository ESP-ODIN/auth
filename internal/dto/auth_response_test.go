package dto

import (
	"encoding/json"
	"testing"
	"time"

	"auth/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserResponse_JSON(t *testing.T) {
	id := uuid.MustParse("3f6c1c1e-8a52-4a8e-9a6b-2f1f0f6a9b10")
	user := &model.User{
		ID:          id,
		Email:       "jane.doe@example.com",
		Username:    "janedoe",
		FirstName:   "Jane",
		LastName:    "Doe",
		DateOfBirth: time.Date(1990, time.January, 31, 0, 0, 0, 0, time.UTC),
		Password:    "$2a$10$hash-qui-ne-doit-jamais-sortir",
		RoleID:      1,
		IsActive:    true,
	}

	raw, err := json.Marshal(NewUserResponse(user))
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(raw, &got))

	// Le mot de passe (même haché) ne doit jamais apparaître dans la réponse
	assert.NotContains(t, got, "password")
	assert.NotContains(t, got, "Password")
	assert.NotContains(t, string(raw), user.Password)

	// Les clés sont en snake_case, comme la requête
	assert.Equal(t, id.String(), got["id"])
	assert.Equal(t, "jane.doe@example.com", got["email"])
	assert.Equal(t, "janedoe", got["username"])
	assert.Equal(t, "Jane", got["first_name"])
	assert.Equal(t, "Doe", got["last_name"])
	assert.Equal(t, float64(1), got["role_id"])
	assert.Equal(t, true, got["is_active"])
	assert.Contains(t, got, "created_at")
	assert.Contains(t, got, "updated_at")

	// La date de naissance est au format YYYY-MM-DD, sans heure
	assert.Equal(t, "1990-01-31", got["date_of_birth"])
}

func TestNewUserResponse_Nil(t *testing.T) {
	assert.Nil(t, NewUserResponse(nil))
}
