package user

import (
	"strings"

	"github.com/google/uuid"
)

// Service handles user business logic.
type Service struct {
	repository *Repository
}

// NewService creates a new user service.
func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// ListUser returns all users.
func (s *Service) ListUser() ([]User, error) {
	return s.repository.ListUser()
}

// GetUser returns a user by ID.
// Returns ErrNotFound for invalid UUID format or when the user does not exist.
func (s *Service) GetUser(id string) (*User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrNotFound
	}
	return s.repository.GetUser(id)
}

// CreateUser creates a new user.
// Email is normalized (lowercase, trimmed) so uniqueness is enforced by email only.
func (s *Service) CreateUser(name, email, password string) (*User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	user := &User{Name: name, Email: normalizedEmail}
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	if err := s.repository.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser updates an existing user by ID. Only name and password can be updated.
func (s *Service) UpdateUser(id string, name *string, password *string) (*User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrNotFound
	}
	user, err := s.repository.GetUser(id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		user.Name = *name
	}
	if password != nil {
		if err := user.SetPassword(*password); err != nil {
			return nil, err
		}
	}
	if err := s.repository.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser soft-deletes a user by ID.
func (s *Service) DeleteUser(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrNotFound
	}
	return s.repository.Delete(id)
}
