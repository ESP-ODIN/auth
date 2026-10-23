package repository

import (
	"auth/internal/model"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}
