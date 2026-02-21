package user

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// En: TokenRevoker is the subset of the auth repository the user service depends on
// to revoke refresh tokens when a user is deleted.
// Es: TokenRevoker es el subconjunto del repositorio de auth del que depende el servicio de usuario
// para revocar los refresh tokens cuando se elimina un usuario.
type TokenRevoker interface {
	RevokeAllByUserID(userID string) error
}

// En: Service handles user business logic.
// Es: Service maneja la lógica de negocio de usuarios.
type Service struct {
	repository   *Repository
	tokenRevoker TokenRevoker
}

// En: NewService creates a new user service.
// Es: NewService crea un nuevo servicio de usuario.
func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// En: WithTokenRevoker sets the token revoker used to invalidate refresh tokens on user deletion.
// Es: WithTokenRevoker establece el revocador de tokens para invalidar refresh tokens al eliminar usuario.
func (service *Service) WithTokenRevoker(tokenRevoker TokenRevoker) *Service {
	service.tokenRevoker = tokenRevoker
	return service
}

// En: GetUser returns a user by ID.
// Returns ErrNotFound for invalid UUID format or when the user does not exist.
// Es: GetUser devuelve un usuario por ID.
// Devuelve ErrNotFound si el UUID es inválido o el usuario no existe.
func (service *Service) GetUser(id string) (*User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrNotFound
	}
	return service.repository.GetUser(id)
}

// En: CreateUser creates a new user.
// Email is normalized (lowercase, trimmed) so uniqueness is enforced by email only.
// Es: CreateUser crea un nuevo usuario.
// El email se normaliza (minúsculas, sin espacios) para que la unicidad sea solo por email.
func (service *Service) CreateUser(name, email, password string) (*User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	user := &User{Name: name, Email: normalizedEmail}
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	if err := service.repository.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

// En: UpdateUser updates an existing user by ID. Only name and password can be updated.
// Es: UpdateUser actualiza un usuario existente por ID. Solo se pueden actualizar nombre y contraseña.
func (service *Service) UpdateUser(id string, name *string, password *string) (*User, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrNotFound
	}
	user, err := service.repository.GetUser(id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		user.Name = *name
	}
	if password != nil {
		if err := user.SetPassword(*password); err != nil {
			return nil, err
		}
	}
	if err := service.repository.Update(user); err != nil {
		return nil, err
	}
	return user, nil
}

// En: DeleteUser soft-deletes a user by ID and revokes all their refresh tokens.
// Es: DeleteUser hace borrado lógico del usuario por ID y revoca todos sus refresh tokens.
func (service *Service) DeleteUser(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrNotFound
	}
	if err := service.repository.Delete(id); err != nil {
		return err
	}
	if service.tokenRevoker != nil {
		if err := service.tokenRevoker.RevokeAllByUserID(id); err != nil {
			return fmt.Errorf("revoke tokens after user delete: %w", err)
		}
	}
	return nil
}
