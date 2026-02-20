package user

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// UpdateMeRequest is the request body for updating the authenticated user's profile.
// At least one field must be present. Email cannot be changed.
type UpdateMeRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=2,max=100"`
	Password *string `json:"password" validate:"omitempty,min=8,max=72"`
}
