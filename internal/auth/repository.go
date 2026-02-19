package auth

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ErrTokenNotFound is returned when a refresh token does not exist.
var ErrTokenNotFound = fmt.Errorf("refresh token not found")

// Repository handles refresh token data access.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create persists a new refresh token.
func (repository *Repository) Create(token *RefreshToken) error {
	if err := repository.db.Create(token).Error; err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

// GetByTokenHash returns a refresh token by its SHA-256 hash.
func (repository *Repository) GetByTokenHash(hash string) (*RefreshToken, error) {
	var token RefreshToken
	if err := repository.db.Where("token_hash = ?", hash).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}
	return &token, nil
}

// Revoke marks a single refresh token as revoked by its ID.
func (repository *Repository) Revoke(id string) error {
	now := time.Now()
	result := repository.db.Model(&RefreshToken{}).Where("id = ?", id).Update("revoked_at", now)
	if result.Error != nil {
		return fmt.Errorf("revoke refresh token: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// RevokeAllByUserID revokes all active refresh tokens for a given user (used on logout).
func (repository *Repository) RevokeAllByUserID(userID string) error {
	now := time.Now()
	if err := repository.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}
