package implementation

import (
	"context"
	"database/sql"
	"auth/internal/model"
)

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM public.user WHERE email = )"
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO public.user (id, email, password, role_id, is_active)
		VALUES (, , , , )
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.RoleID, user.IsActive)
	return err
}
