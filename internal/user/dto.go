package user

// En: CreateUserRequest is the request body for creating a user.
// Es: CreateUserRequest es el cuerpo de la petición para crear un usuario.
type CreateUserRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// En: UpdateMeRequest is the request body for updating the authenticated user's profile.
// At least one field must be present. Email cannot be changed.
// Es: UpdateMeRequest es el cuerpo de la petición para actualizar el perfil del usuario autenticado.
// Debe estar presente al menos un campo. El email no se puede cambiar.
type UpdateMeRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=2,max=100"`
	Password *string `json:"password" validate:"omitempty,min=8,max=72"`
}
