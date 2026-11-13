package implementation

import (
	"auth/internal/errors"
	"auth/internal/model"
	"context"
	stdErrors "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM public.user WHERE LOWER(email) = LOWER($1))"
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepositoryImpl) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO public.user (id, email, password, username, first_name, last_name, date_of_birth, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`
	// created_at et updated_at sont remplis par la base : on les relit pour que la réponse soit exacte
	return r.db.QueryRow(ctx, query, user.ID, user.Email, user.Password, user.Username, user.FirstName, user.LastName, user.DateOfBirth, user.RoleID, user.IsActive).
		Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM public.user WHERE LOWER(username) = LOWER($1))"
	err := r.db.QueryRow(ctx, query, username).Scan(&exists)
	return exists, err
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := "SELECT id, email, password, role_id, is_active, created_at, updated_at FROM public.user WHERE LOWER(email) = LOWER($1)"
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.ErrorUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	query := "SELECT id, email, password, role_id, is_active, created_at, updated_at FROM public.user WHERE id = $1"
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email, &user.Password, &user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if stdErrors.Is(err, pgx.ErrNoRows) {
			return nil, errors.ErrorUserNotFound
		}
		return nil, err
	}
	return &user, nil
}