package user

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
func (s *Service) GetUser(id string) (*User, error) {
	return s.repository.GetUser(id)
}
