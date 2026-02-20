package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const bcryptCost = 12

// User represents a user in the system.
type User struct {
	ID                         string         `gorm:"type:uuid;primaryKey" json:"id"`
	Name                       string         `gorm:"not null" json:"name"`
	Email                      string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash               string         `gorm:"column:password_hash;not null" json:"-"`
	EmailVerifiedAt            *time.Time     `gorm:"column:email_verified_at" json:"email_verified_at,omitempty"`
	EmailVerificationToken     *string        `gorm:"column:email_verification_token;index" json:"-"`
	EmailVerificationExpiresAt *time.Time     `gorm:"column:email_verification_expires_at" json:"-"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	DeletedAt                  gorm.DeletedAt `gorm:"index" json:"-"`
}

// IsEmailVerified returns true if the user has verified their email address.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// SetPassword hashes the plain password and stores it in PasswordHash.
func (u *User) SetPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies that the plain password matches the stored hash.
func (u *User) CheckPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)) == nil
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}

// BeforeCreate generates UUID before insert.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
