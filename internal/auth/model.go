package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProviderType identifies the authentication provider.
type ProviderType string

const (
	ProviderCredentials ProviderType = "credentials"
	ProviderGoogle      ProviderType = "google"
	ProviderFacebook    ProviderType = "facebook"
)

// UserAuthProvider links a User to an external authentication provider.
// UNIQUE(provider, provider_subject_id) ensures one identity per provider slot.
type UserAuthProvider struct {
	ID                string       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            string       `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider          ProviderType `gorm:"not null" json:"provider"`
	ProviderSubjectID string       `gorm:"column:provider_subject_id;not null" json:"provider_subject_id"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

// TableName overrides the table name.
func (UserAuthProvider) TableName() string {
	return "user_auth_providers"
}

// BeforeCreate generates UUID before insert.
func (p *UserAuthProvider) BeforeCreate(_ *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// RefreshToken represents a stored refresh token tied to a user session.
// The raw token is never persisted; only its SHA-256 hash is stored.
type RefreshToken struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"-"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"-"`
	TokenHash string         `gorm:"column:token_hash;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"-"`
	RevokedAt *time.Time     `gorm:"index" json:"-"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// BeforeCreate generates UUID before insert.
func (rt *RefreshToken) BeforeCreate(_ *gorm.DB) error {
	if rt.ID == "" {
		rt.ID = uuid.New().String()
	}
	return nil
}

// IsExpired returns true if the token has passed its expiry time.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked returns true if the token has been explicitly revoked.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
