package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
