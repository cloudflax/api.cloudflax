package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Post representa un post/publicación de un usuario.
type Post struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	Title     string         `gorm:"not null" json:"title"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName sobrescribe el nombre de la tabla.
func (Post) TableName() string {
	return "posts"
}

// BeforeCreate genera UUID antes de insertar.
func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
