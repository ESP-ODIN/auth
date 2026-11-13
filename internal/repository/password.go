package repository

import "auth/internal/model"

type PasswordRepository interface {
	FindByID(userID string) (*model.User, error)
}