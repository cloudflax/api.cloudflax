package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Repository handles user data access.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ErrNotFound is returned when a user is not found.
var ErrNotFound = fmt.Errorf("user not found")

// ErrDuplicateEmail is returned when creating a user with an email that already exists.
var ErrDuplicateEmail = fmt.Errorf("email already exists")

// GetUserByEmail returns a user by email address.
func (repository *Repository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := repository.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// GetUser returns a user by ID.
func (repository *Repository) GetUser(id string) (*User, error) {
	var user User
	if err := repository.db.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

// ExistsByEmail returns true if a user with the given email exists, optionally excluding an ID (for updates).
// Includes soft-deleted users so the same email cannot be reused after delete.
func (repository *Repository) ExistsByEmail(email, excludeID string) (bool, error) {
	var count int64
	query := repository.db.Unscoped().Model(&User{}).Where("email = ?", email)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("exists by email: %w", err)
	}
	return count > 0, nil
}

// Create creates a new user.
func (repository *Repository) Create(user *User) error {
	exists, err := repository.ExistsByEmail(user.Email, "")
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}
	if err := repository.db.Create(user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// Update updates an existing user.
func (repository *Repository) Update(user *User) error {
	if err := repository.db.Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// Delete soft-deletes a user by ID.
func (repository *Repository) Delete(id string) error {
	result := repository.db.Where("id = ?", id).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
