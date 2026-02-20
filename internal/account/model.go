package account

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleType identifies the role of a member within an account.
type RoleType string

const (
	RoleOwner  RoleType = "owner"
	RoleAdmin  RoleType = "admin"
	RoleMember RoleType = "member"
)

// Account represents an organization or workspace in the system.
type Account struct {
	ID        string         `gorm:"type:uuid;primaryKey"    json:"id"`
	Name      string         `gorm:"not null"                json:"name"`
	Slug      string         `gorm:"uniqueIndex;not null"    json:"slug"`
	CreatedAt time.Time      `                               json:"created_at"`
	UpdatedAt time.Time      `                               json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                   json:"-"`
}

// TableName overrides the table name.
func (Account) TableName() string {
	return "accounts"
}

// BeforeCreate generates UUID before insert.
func (a *Account) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// AccountMember links a User to an Account with a specific role.
// UNIQUE(account_id, user_id) ensures one membership per user per account.
type AccountMember struct {
	ID        string    `gorm:"type:uuid;primaryKey"                          json:"id"`
	AccountID string    `gorm:"type:uuid;not null;uniqueIndex:idx_account_user" json:"account_id"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_account_user" json:"user_id"`
	Role      RoleType  `gorm:"not null;default:'member'"                     json:"role"`
	CreatedAt time.Time `                                                     json:"created_at"`
	UpdatedAt time.Time `                                                     json:"updated_at"`
}

// TableName overrides the table name.
func (AccountMember) TableName() string {
	return "account_members"
}

// BeforeCreate generates UUID before insert.
func (m *AccountMember) BeforeCreate(_ *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}
