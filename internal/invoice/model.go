package invoice

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatusType represents the lifecycle state of an invoice.
type StatusType string

const (
	StatusDraft  StatusType = "draft"
	StatusSent   StatusType = "sent"
	StatusPaid   StatusType = "paid"
	StatusVoided StatusType = "voided"
)

// Invoice represents a billing document scoped to an account.
type Invoice struct {
	ID         string         `gorm:"type:uuid;primaryKey"     json:"id"`
	AccountID  string         `gorm:"type:uuid;not null;index"  json:"account_id"`
	Number     string         `gorm:"not null"                  json:"number"`
	Status     StatusType     `gorm:"not null;default:'draft'"  json:"status"`
	TotalCents int64          `gorm:"not null;default:0"        json:"total_cents"`
	Currency   string         `gorm:"not null;default:'USD'"    json:"currency"`
	IssuedAt   *time.Time     `                                 json:"issued_at,omitempty"`
	DueAt      *time.Time     `                                 json:"due_at,omitempty"`
	CreatedAt  time.Time      `                                 json:"created_at"`
	UpdatedAt  time.Time      `                                 json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                     json:"-"`
}

// TableName overrides the table name.
func (Invoice) TableName() string {
	return "invoices"
}

// BeforeCreate generates a UUID before insert.
func (invoice *Invoice) BeforeCreate(_ *gorm.DB) error {
	if invoice.ID == "" {
		invoice.ID = uuid.New().String()
	}
	return nil
}
