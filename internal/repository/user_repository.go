package repository

import model "auth/internal/model"

type UserRepo interface {
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
}
