package auth

// En: RegisterRequest is the request body for POST /auth/register.
// Es: Request body para POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// En: VerifyEmailRequest represents the query parameters for the email verification endpoint (/auth/verify-email).
// Es: VerifyEmailRequest representa los parámetros de consulta para el endpoint de verificación de correo electrónico (/auth/verify-email).
type VerifyEmailRequest struct {
	Token string `query:"token" validate:"required"`
}

// En: ResendVerificationRequest represents the request body for the email verification resend endpoint.
// Es: ResendVerificationRequest representa el cuerpo de la solicitud para el endpoint de reenvío de verificación de correo electrónico.
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// En: LoginRequest represents the request body for the login endpoint.
// Es: LoginRequest representa el cuerpo de la solicitud para el endpoint de inicio de sesión.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// En: RefreshRequest represents the request body for the refresh endpoint.
// Es: RefreshRequest representa el cuerpo de la solicitud para el endpoint de actualización.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
