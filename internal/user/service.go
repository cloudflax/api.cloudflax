package user

// Service handles user business logic.
type Service struct {
	repo *Repository
}

// NewService creates a new user service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListUser returns all users.
func (s *Service) ListUser() ([]User, error) {
	return s.repo.ListUser()
}

// GetUser returns a user by ID.
func (s *Service) GetUser(id string) (*User, error) {
	return s.repo.GetUser(id)
}
