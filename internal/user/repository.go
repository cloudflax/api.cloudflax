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
func (r *Repository) ListUser() ([]User, error) {
	var users []User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

// GetUser returns a user by ID.
func (r *Repository) GetUser(id string) (*User, error) {
	var u User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

// Create creates a new user.
func (r *Repository) Create(u *User) error {
	if err := r.db.Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}
