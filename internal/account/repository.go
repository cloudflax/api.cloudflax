package account

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Repository handles account and account_member data access.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new account repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateAccount persists a new account.
// Returns ErrSlugTaken if the slug is already used by another account.
func (r *Repository) CreateAccount(account *Account) error {
	if err := r.db.Create(account).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrSlugTaken
		}
		return fmt.Errorf("create account: %w", err)
	}
	return nil
}

// GetByID returns an account by its primary key.
func (r *Repository) GetByID(id string) (*Account, error) {
	var account Account
	if err := r.db.First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get account by id: %w", err)
	}
	return &account, nil
}

// GetBySlug returns an account by its unique slug.
func (r *Repository) GetBySlug(slug string) (*Account, error) {
	var account Account
	if err := r.db.First(&account, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get account by slug: %w", err)
	}
	return &account, nil
}

// SlugExists returns true if any account (including soft-deleted) uses the given slug.
func (r *Repository) SlugExists(slug string) (bool, error) {
	var count int64
	if err := r.db.Unscoped().Model(&Account{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, fmt.Errorf("slug exists: %w", err)
	}
	return count > 0, nil
}

// CreateMember persists a new account membership.
func (r *Repository) CreateMember(member *AccountMember) error {
	if err := r.db.Create(member).Error; err != nil {
		return fmt.Errorf("create account member: %w", err)
	}
	return nil
}

// GetMember returns the membership for a given account and user.
// Returns ErrMemberNotFound when no such membership exists.
func (r *Repository) GetMember(accountID, userID string) (*AccountMember, error) {
	var member AccountMember
	if err := r.db.First(&member, "account_id = ? AND user_id = ?", accountID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMemberNotFound
		}
		return nil, fmt.Errorf("get account member: %w", err)
	}
	return &member, nil
}

// ListMembers returns all memberships for the given account.
func (r *Repository) ListMembers(accountID string) ([]AccountMember, error) {
	var members []AccountMember
	if err := r.db.Where("account_id = ?", accountID).Find(&members).Error; err != nil {
		return nil, fmt.Errorf("list account members: %w", err)
	}
	return members, nil
}

// ListAccountsForUser returns all accounts where the given user is a member.
func (r *Repository) ListAccountsForUser(userID string) ([]Account, error) {
	var accounts []Account
	if err := r.db.
		Joins("JOIN account_members ON account_members.account_id = accounts.id").
		Where("account_members.user_id = ?", userID).
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("list accounts for user: %w", err)
	}
	return accounts, nil
}
