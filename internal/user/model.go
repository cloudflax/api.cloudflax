package user

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const bcryptCost = 12

// En: User represents a user in the system.
// Es: User representa un usuario en el sistema.
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

// En: IsEmailVerified returns true if the user has verified their email address.
// Es: IsEmailVerified devuelve true si el usuario ha verificado su dirección de email.
func (user *User) IsEmailVerified() bool {
	return user.EmailVerifiedAt != nil
}

// En: SetPassword hashes the plain password and stores it in PasswordHash.
// Es: SetPassword hashea la contraseña en claro y la guarda en PasswordHash.
func (user *User) SetPassword(plain string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	return nil
}

// En: CheckPassword verifies that the plain password matches the stored hash.
// Es: CheckPassword verifica que la contraseña en claro coincida con el hash almacenado.
func (user *User) CheckPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plain)) == nil
}

// En: TableName overrides the table name.
// Es: TableName sobrescribe el nombre de la tabla.
func (User) TableName() string {
	return "users"
}

// En: BeforeCreate generates UUID before insert.
// Es: BeforeCreate genera el UUID antes del insert.
func (user *User) BeforeCreate(tx *gorm.DB) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	return nil
}
