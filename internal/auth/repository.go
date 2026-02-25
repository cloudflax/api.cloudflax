package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/user"
	"gorm.io/gorm"
)

// En: Repository handles refresh token data access.
// Es: Repository maneja el acceso a los datos de los tokens de actualización.
type Repository struct {
	db *gorm.DB
}

// En: NewRepository creates a new auth repository.
// Es: NewRepository crea un nuevo repositorio de autenticación.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// En: Create persists a new refresh token.
// Es: Create persiste un nuevo token de actualización.
func (repository *Repository) Create(token *RefreshToken) error {
	if err := repository.db.Create(token).Error; err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

// En: GetByTokenHash returns a refresh token by its SHA-256 hash.
// Es: GetByTokenHash devuelve un token de actualización por su hash SHA-256.
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

// En: Revoke marks a single refresh token as revoked by its ID.
// Es: Revoca un token de actualización por su ID.
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

// En: RevokeAllByUserID revokes all active refresh tokens for a given user (used on logout).
// Es: Revoca todos los tokens de actualización activos para un usuario dado (usado en el cierre de sesión).
func (repository *Repository) RevokeAllByUserID(userID string) error {
	now := time.Now()
	if err := repository.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("revoke all user tokens: %w", err)
	}
	return nil
}

// En: CreateAuthProvider persists a new UserAuthProvider record.
// Es: CreateAuthProvider persiste un nuevo registro de UserAuthProvider.
func (repository *Repository) CreateAuthProvider(provider *UserAuthProvider) error {
	if err := repository.db.Create(provider).Error; err != nil {
		return fmt.Errorf("create auth provider: %w", err)
	}
	return nil
}

// En: FindByProviderAndSubject returns the UserAuthProvider matching the given provider type and subject ID.
// Es: FindByProviderAndSubject devuelve el UserAuthProvider que coincide con el tipo de proveedor y el ID de sujeto dado.
func (repository *Repository) FindByProviderAndSubject(provider ProviderType, subjectID string) (*UserAuthProvider, error) {
	var p UserAuthProvider
	err := repository.db.
		Where("provider = ? AND provider_subject_id = ?", provider, subjectID).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("find auth provider: %w", err)
	}
	return &p, nil
}

// En: FindByVerificationToken returns the user that owns the given email verification token.
// Es: FindByVerificationToken devuelve el usuario que posee el token de verificación de correo electrónico dado.
func (repository *Repository) FindByVerificationToken(token string) (*user.User, error) {
	var u user.User
	err := repository.db.Where("email_verification_token = ?", token).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidVerificationToken
		}
		return nil, fmt.Errorf("find by verification token: %w", err)
	}
	return &u, nil
}
