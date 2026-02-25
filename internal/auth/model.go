// Package auth provides authentication: registration with email verification,
// login/refresh token pairs, JWT access tokens, and logout (revoke refresh tokens).
package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// En: ProviderType identifies the authentication provider.
// Es: ProviderType identifica el proveedor de autenticación.
type ProviderType string

const (
	ProviderCredentials ProviderType = "credentials"
	ProviderGoogle      ProviderType = "google"
	ProviderFacebook    ProviderType = "facebook"
)

// En: UserAuthProvider links a User to an external authentication provider.
// Es: UserAuthProvider enlaza un Usuario a un proveedor de autenticación externo.
type UserAuthProvider struct {
	ID                string       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            string       `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider          ProviderType `gorm:"not null" json:"provider"`
	ProviderSubjectID string       `gorm:"column:provider_subject_id;not null" json:"provider_subject_id"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

// En: TableName overrides the table name.
// Es: TableName sobrescribe el nombre de la tabla.
func (UserAuthProvider) TableName() string {
	return "user_auth_providers"
}

// En: BeforeCreate generates UUID before insert.
// Es: BeforeCreate genera UUID antes de insertar.
func (p *UserAuthProvider) BeforeCreate(_ *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

// En: RefreshToken represents a stored refresh token tied to a user session.
// Es: RefreshToken representa un token de actualización almacenado asociado a una sesión de usuario.
type RefreshToken struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"-"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"-"`
	TokenHash string         `gorm:"column:token_hash;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time      `gorm:"not null" json:"-"`
	RevokedAt *time.Time     `gorm:"index" json:"-"`
	CreatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// En: TableName overrides the table name.
// Es: TableName sobrescribe el nombre de la tabla.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// En: BeforeCreate generates UUID before insert.
// Es: BeforeCreate genera UUID antes de insertar.
func (rt *RefreshToken) BeforeCreate(_ *gorm.DB) error {
	if rt.ID == "" {
		rt.ID = uuid.New().String()
	}
	return nil
}

// En: IsExpired returns true if the token has passed its expiry time.
// Es: IsExpired devuelve true si el token ha pasado su tiempo de expiración.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// En: IsRevoked returns true if the token has been explicitly revoked.
// Es: IsRevoked devuelve true si el token ha sido revocado explícitamente.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
