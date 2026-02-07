package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User representa un usuario en el sistema.
type User struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName sobrescribe el nombre de la tabla.
func (User) TableName() string {
	return "users"
}

// BeforeCreate genera UUID antes de insertar.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
