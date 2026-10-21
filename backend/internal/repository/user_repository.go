package repository

import (
	"context"
	"fmt"

	"github.com/GlitchOfTheMatrix/Todo-App/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).
		Scan(&user.Id, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).
		Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).
		Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}
