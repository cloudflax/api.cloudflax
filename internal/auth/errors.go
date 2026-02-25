package auth

import "fmt"

// En: ErrInvalidCredentials is returned when login credentials are wrong or refresh token is invalid/expired.
// Es: ErrInvalidCredentials se devuelve cuando las credenciales de inicio de sesión son incorrectas o el token de actualización es inválido/expirado.
var ErrInvalidCredentials = fmt.Errorf("invalid credentials")

// En: ErrInvalidVerificationToken is returned when the email verification token is invalid or expired.
// Es: ErrInvalidVerificationToken se devuelve cuando el token de verificación de correo electrónico es inválido o expirado.
var ErrInvalidVerificationToken = fmt.Errorf("invalid verification token")

// En: ErrEmailAlreadyVerified is returned when the email is already verified.
// Es: ErrEmailAlreadyVerified se devuelve cuando el correo electrónico ya está verificado.
var ErrEmailAlreadyVerified = fmt.Errorf("email already verified")

// En: ErrEmailNotVerified is returned when login or refresh is attempted with an unverified email.
// Es: ErrEmailNotVerified se devuelve cuando se intenta iniciar sesión o actualizar con un correo electrónico no verificado.
var ErrEmailNotVerified = fmt.Errorf("email not verified")

// En: ErrTokenNotFound is returned when a refresh token does not exist.
// Es: ErrTokenNotFound se devuelve cuando un token de actualización no existe.
var ErrTokenNotFound = fmt.Errorf("refresh token not found")
