package dto

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

// validRegisterRequest renvoie une requête qui respecte toutes les règles du DTO.
func validRegisterRequest() RegisterRequest {
	return RegisterRequest{
		Email:       "jane.doe@example.com",
		Password:    "Str0ngPassw0rd!",
		Username:    "janedoe",
		FirstName:   "Jane",
		LastName:    "Doe",
		DateOfBirth: "1990-01-31",
	}
}

// TestRegisterRequest_Validation vérifie les tags `binding` du DTO avec le même
// validateur que Gin utilise dans le handler (c.ShouldBindJSON).
func TestRegisterRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(r *RegisterRequest)
		wantErr bool
	}{
		{"requête valide", func(r *RegisterRequest) {}, false},

		// email
		{"email manquant", func(r *RegisterRequest) { r.Email = "" }, true},
		{"email invalide", func(r *RegisterRequest) { r.Email = "pas-un-email" }, true},

		// password
		{"mot de passe manquant", func(r *RegisterRequest) { r.Password = "" }, true},
		{"mot de passe trop court (7)", func(r *RegisterRequest) { r.Password = "1234567" }, true},
		{"mot de passe longueur min (8)", func(r *RegisterRequest) { r.Password = "12345678" }, false},
		{"mot de passe longueur max (72)", func(r *RegisterRequest) { r.Password = strings.Repeat("a", 72) }, false},
		{"mot de passe trop long (73)", func(r *RegisterRequest) { r.Password = strings.Repeat("a", 73) }, true},

		// username
		{"pseudo manquant", func(r *RegisterRequest) { r.Username = "" }, true},
		{"pseudo trop court (2)", func(r *RegisterRequest) { r.Username = "ab" }, true},
		{"pseudo longueur min (3)", func(r *RegisterRequest) { r.Username = "abc" }, false},
		{"pseudo longueur max (50)", func(r *RegisterRequest) { r.Username = strings.Repeat("a", 50) }, false},
		{"pseudo trop long (51)", func(r *RegisterRequest) { r.Username = strings.Repeat("a", 51) }, true},

		// first_name
		{"prénom manquant", func(r *RegisterRequest) { r.FirstName = "" }, true},
		{"prénom trop court (2)", func(r *RegisterRequest) { r.FirstName = "Jo" }, true},
		{"prénom longueur min (3)", func(r *RegisterRequest) { r.FirstName = "Ana" }, false},
		{"prénom trop long (101)", func(r *RegisterRequest) { r.FirstName = strings.Repeat("a", 101) }, true},

		// last_name
		{"nom manquant", func(r *RegisterRequest) { r.LastName = "" }, true},
		{"nom trop court (1)", func(r *RegisterRequest) { r.LastName = "D" }, true},
		{"nom longueur min (2)", func(r *RegisterRequest) { r.LastName = "Li" }, false},
		{"nom trop long (101)", func(r *RegisterRequest) { r.LastName = strings.Repeat("a", 101) }, true},

		// date_of_birth
		{"date de naissance manquante", func(r *RegisterRequest) { r.DateOfBirth = "" }, true},
		{"date au format français", func(r *RegisterRequest) { r.DateOfBirth = "31/01/1990" }, true},
		{"date avec l'heure", func(r *RegisterRequest) { r.DateOfBirth = "1990-01-31T10:00:00Z" }, true},
		{"date inexistante", func(r *RegisterRequest) { r.DateOfBirth = "1990-02-30" }, true},
		{"date sans zéros", func(r *RegisterRequest) { r.DateOfBirth = "1990-1-31" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validRegisterRequest()
			tt.modify(&req)

			err := binding.Validator.ValidateStruct(req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
