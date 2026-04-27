package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"convoy/internal/models"
	"convoy/internal/repository"

	"github.com/google/uuid"
)

type userRepository struct {
	db *DB
}

func NewUserRepository(db *DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, email, password_hash, display_name, profile_pic, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.ProfilePic,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, email, password_hash, display_name, profile_pic, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, email, password_hash, display_name, profile_pic, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = $1, password_hash = $2, display_name = $3, profile_pic = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.Exec(
		query,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.ProfilePic,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
