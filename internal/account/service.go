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
	Update(user *user.User) error
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

// ListAccountsForUser returns all accounts where the given user is a member.
// Returns user.ErrNotFound when the user ID is not a valid UUID.
func (s *Service) ListAccountsForUser(userID string) ([]Account, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, user.ErrNotFound
	}
	return s.repository.ListAccountsForUser(userID)
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

	// If the owner does not have an active account yet, set this one as active.
	if u.ActiveAccountID == nil {
		accountID := account.ID
		u.ActiveAccountID = &accountID
		if err := s.userRepository.Update(u); err != nil {
			return nil, nil, fmt.Errorf("set active account for owner: %w", err)
		}
	}

	return account, member, nil
}

// SetActiveAccountForUser marks the given account as the active account for the given user.
// It validates UUID formats, ensures the account exists and that the user is a member of it.
// Returns user.ErrNotFound when the user ID is invalid or the user does not exist,
// ErrNotFound when the account does not exist and ErrMemberNotFound when the user is not a member.
func (s *Service) SetActiveAccountForUser(userID, accountID string) (*user.User, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, user.ErrNotFound
	}
	if _, err := uuid.Parse(accountID); err != nil {
		return nil, ErrNotFound
	}

	if _, err := s.repository.GetMember(accountID, userID); err != nil {
		return nil, err
	}

	u, err := s.userRepository.GetUser(userID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}

	u.ActiveAccountID = &accountID
	if err := s.userRepository.Update(u); err != nil {
		return nil, fmt.Errorf("update active account: %w", err)
	}

	return u, nil
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
