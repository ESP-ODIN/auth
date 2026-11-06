package dto

import (
	"time"

	"auth/internal/model"
)

// DateLayout est le format des dates sans heure exposées par l'API (YYYY-MM-DD).
const DateLayout = "2006-01-02"

// UserResponse est la représentation publique d'un utilisateur renvoyée par l'API.
// Elle ne contient jamais le mot de passe (même haché).
type UserResponse struct {
	ID          string    `json:"id"            example:"3f6c1c1e-8a52-4a8e-9a6b-2f1f0f6a9b10"`
	Email       string    `json:"email"         example:"jane.doe@example.com"`
	Username    string    `json:"username"      example:"janedoe"`
	FirstName   string    `json:"first_name"    example:"Jane"`
	LastName    string    `json:"last_name"     example:"Doe"`
	DateOfBirth string    `json:"date_of_birth" example:"1990-01-31" format:"date"`
	RoleID      int       `json:"role_id"       example:"1"`
	IsActive    bool      `json:"is_active"     example:"true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewUserResponse convertit l'entité model.User en réponse publique.
func NewUserResponse(user *model.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth.Format(DateLayout),
		RoleID:      user.RoleID,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

type RegisterResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

type LoginResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}
