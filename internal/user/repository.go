package user

import (
	"fmt"

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

// ListUser returns all users.
func (repository *Repository) ListUser() ([]User, error) {
	var users []User
	if err := repository.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
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

// Create creates a new user.
func (repository *Repository) Create(user *User) error {
	if err := repository.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
