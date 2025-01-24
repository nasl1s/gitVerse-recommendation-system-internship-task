package repository

import (
	"context"
	"fmt"
	"recommendation-system/internal/sso/models"
	"recommendation-system/pkg/db"
	log "recommendation-system/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.UserSSO) error
	GetUserByEmail(ctx context.Context, email string) (*models.UserSSO, error)
}

type userRepository struct {
	db     *db.DB
	logger *log.Logger
}

func NewUserRepository(database *db.DB, logger *log.Logger) UserRepository {
	return &userRepository{db: database, logger: logger}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.UserSSO) error {
	r.logger.Println("Creating new user")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		r.logger.Printf("Failed to hash password: %v", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	r.logger.Printf("Executing query to insert user: %s", user.Email)
	err = r.db.Pool.QueryRow(ctx, query, user.Name, user.Email, string(hashedPassword)).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.logger.Printf("Failed to create user: %v", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.PasswordHash = string(hashedPassword)
	r.logger.Printf("User created successfully with ID: %d", user.ID)
	return nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.UserSSO, error) {
	r.logger.Printf("Fetching user by email: %s", email)
	var user models.UserSSO
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.logger.Printf("Failed to get user by email: %v", err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	r.logger.Printf("User fetched successfully with ID: %d", user.ID)
	return &user, nil
}
