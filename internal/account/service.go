package account

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/google/uuid"
)

// UserRepository is the subset of the user repository the account service depends on.
type UserRepository interface {
	GetUser(id string) (*user.User, error)
}

// Service handles account business logic.
type Service struct {
	repository     *Repository
	userRepository UserRepository
}

// NewService creates a new account service.
func NewService(repository *Repository, userRepository UserRepository) *Service {
	return &Service{repository: repository, userRepository: userRepository}
}

// CreateAccount creates a new account owned by the given user.
// If slug is empty it is derived from name. Returns ErrUserEmailNotVerified when the
// owner's email has not been verified, and ErrSlugTaken when the slug is already in use.
func (s *Service) CreateAccount(name, slug, ownerUserID string) (*Account, *AccountMember, error) {
	if _, err := uuid.Parse(ownerUserID); err != nil {
		return nil, nil, user.ErrNotFound
	}

	u, err := s.userRepository.GetUser(ownerUserID)
	if err != nil {
		return nil, nil, fmt.Errorf("lookup owner: %w", err)
	}
	if !u.IsEmailVerified() {
		return nil, nil, ErrUserEmailNotVerified
	}

	if slug == "" {
		slug = slugify(name)
	}

	taken, err := s.repository.SlugExists(slug)
	if err != nil {
		return nil, nil, fmt.Errorf("check slug: %w", err)
	}
	if taken {
		return nil, nil, ErrSlugTaken
	}

	account := &Account{Name: name, Slug: slug}
	if err := s.repository.CreateAccount(account); err != nil {
		return nil, nil, err
	}

	member := &AccountMember{
		AccountID: account.ID,
		UserID:    ownerUserID,
		Role:      RoleOwner,
	}
	if err := s.repository.CreateMember(member); err != nil {
		return nil, nil, fmt.Errorf("create owner member: %w", err)
	}

	return account, member, nil
}

var (
	nonAlphanumDash = regexp.MustCompile(`[^a-z0-9-]+`)
	multipleDashes  = regexp.MustCompile(`-{2,}`)
)

// slugify converts a name into a URL-safe slug (lowercase, words joined by dashes).
func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlphanumDash.ReplaceAllString(s, "")
	s = multipleDashes.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = uuid.New().String()
	}
	return s
}
