package user

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// En: Repository handles user data access.
// Es: Repository maneja el acceso a datos de usuarios.
type Repository struct {
	db *gorm.DB
}

// En: NewRepository creates a new user repository.
// Es: NewRepository crea un nuevo repositorio de usuario.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// En: GetUserByEmail returns a user by email address.
// Es: GetUserByEmail devuelve un usuario por dirección de email.
func (repository *Repository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := repository.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// En: GetUser returns a user by ID.
// Es: GetUser devuelve un usuario por ID.
func (repository *Repository) GetUser(id string) (*User, error) {
	var user User
	if err := repository.db.First(&user, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

// En: ExistsByEmail returns true if a user with the given email exists, optionally excluding an ID (for updates).
// Includes soft-deleted users so the same email cannot be reused after delete.
// Es: ExistsByEmail devuelve true si existe un usuario con el email dado, opcionalmente excluyendo un ID (para actualizaciones).
// Incluye usuarios con borrado lógico para que el mismo email no se pueda reutilizar tras borrar.
func (repository *Repository) ExistsByEmail(email, excludeID string) (bool, error) {
	var count int64
	query := repository.db.Unscoped().Model(&User{}).Where("email = ?", email)
	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("exists by email: %w", err)
	}
	return count > 0, nil
}

// En: Create creates a new user.
// Es: Create crea un nuevo usuario.
func (repository *Repository) Create(user *User) error {
	exists, err := repository.ExistsByEmail(user.Email, "")
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}
	if err := repository.db.Create(user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// En: Update updates an existing user.
// Es: Update actualiza un usuario existente.
func (repository *Repository) Update(user *User) error {
	if err := repository.db.Save(user).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// En: Delete soft-deletes a user by ID.
// Es: Delete hace borrado lógico de un usuario por ID.
func (repository *Repository) Delete(id string) error {
	result := repository.db.Where("id = ?", id).Delete(&User{})
	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
