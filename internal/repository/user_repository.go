package repository

import (
	"auth/internal/model"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}
