package dto

import "auth/internal/model"

type RegisterResponse struct {
	User  *model.User `json:"user"`
	Token string      `json:"token"`
}
