package auth

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/cloudflax/api.cloudflax/internal/shared/requestctx"
	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/shared/validator"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
)

// En: Handler handles HTTP requests for authentication.
// Es: Manejador maneja las solicitudes HTTP para la autenticación.
type Handler struct {
	service      *Service
	resendGuard  ResendVerificationGuard
	forgotGuard  ForgotPasswordGuard
}

// En: NewHandler creates a new auth handler.
// Es: Crea un nuevo manejador de autenticación.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// En: WithResendVerificationGuard sets an optional resend verification guard.
// Es: WithResendVerificationGuard define un guard opcional para reenvio.
func (handler *Handler) WithResendVerificationGuard(guard ResendVerificationGuard) *Handler {
	handler.resendGuard = guard
	return handler
}

// En: WithForgotPasswordGuard sets an optional forgot-password email throttle guard.
// Es: WithForgotPasswordGuard define un guard opcional de throttle para correo de recuperacion.
func (handler *Handler) WithForgotPasswordGuard(guard ForgotPasswordGuard) *Handler {
	handler.forgotGuard = guard
	return handler
}

// En: Login authenticates a user and returns an access + refresh token pair.
// Es: Inicia sesión de un usuario y devuelve un par de tokens de acceso y actualización.
func (handler *Handler) Login(ctx fiber.Ctx) error {
	var req LoginRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("login bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("login validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	pair, err := handler.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return runtimeError.Respond(ctx, fiber.StatusUnauthorized, runtimeError.CodeInvalidCredentials, "Invalid email or password")
		}
		if errors.Is(err, ErrEmailNotVerified) {
			return runtimeError.Respond(ctx, fiber.StatusForbidden, runtimeError.CodeEmailVerificationRequired, "Email must be verified before you can log in")
		}
		slog.Error("login", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Login failed")
	}

	return ctx.JSON(fiber.Map{"data": pair})
}

// En: Refresh exchanges a valid refresh token for a new token pair.
// Es: Actualiza un token de actualización válido por un nuevo par de tokens.
func (handler *Handler) Refresh(ctx fiber.Ctx) error {
	var req RefreshRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("refresh bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	pair, err := handler.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrJWTUsedAsRefreshToken) {
			return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeRefreshTokenWrongFormat,
				"Use refresh_token (opaque value from login), not access_token (JWT)")
		}
		if errors.Is(err, ErrInvalidCredentials) {
			return runtimeError.Respond(ctx, fiber.StatusUnauthorized, runtimeError.CodeTokenInvalid, "Invalid or expired refresh token")
		}
		if errors.Is(err, ErrEmailNotVerified) {
			return runtimeError.Respond(ctx, fiber.StatusForbidden, runtimeError.CodeEmailVerificationRequired, "Email must be verified before you can use this session")
		}
		slog.Error("refresh tokens", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Token refresh failed")
	}

	return ctx.JSON(fiber.Map{"data": pair})
}

// En: Logout revokes all active refresh tokens for the authenticated user.
// Es: Cierra sesión de un usuario y revoca todos los tokens de actualización activos.
func (handler *Handler) Logout(ctx fiber.Ctx) error {
	requestContext, err := requestctx.UserOnly(ctx)
	if err != nil {
		return runtimeError.Respond(ctx, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Unauthorized")
	}

	if err := handler.service.Logout(requestContext.UserID); err != nil {
		slog.Error("logout", "user_id", requestContext.UserID, "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Logout failed")
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

// En: Register creates a new user account with email/password credentials.
// Es: Crea una nueva cuenta de usuario con credenciales de email/contraseña.
func (handler *Handler) Register(ctx fiber.Ctx) error {
	var req RegisterRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("register bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("register validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	createdUser, _, err := handler.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			return runtimeError.Respond(ctx, fiber.StatusConflict, runtimeError.CodeEmailAlreadyExists, "Email already exists")
		}
		slog.Error("register", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Registration failed")
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": createdUser,
		"meta": fiber.Map{"email_verification_required": true},
	})
}

// En: VerifyEmail marks the user's email as verified using the token from the verification link.
// Es: Marca el correo electrónico del usuario como verificado usando el token del enlace de verificación.
func (handler *Handler) VerifyEmail(ctx fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := ctx.Bind().Query(&req); err != nil {
		slog.Debug("verify email bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request")
	}
	if err := validator.Validate(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	if err := handler.service.VerifyEmail(req.Token); err != nil {
		if errors.Is(err, ErrInvalidVerificationToken) {
			return runtimeError.Respond(ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeInvalidVerificationToken, "Invalid or expired verification token")
		}
		slog.Error("verify email", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Email verification failed")
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Email verified successfully"})
}

// En: ResendVerification generates a new verification token for the given email.
// Es: Envía un nuevo correo de verificación para el correo electrónico dado.
func (handler *Handler) ResendVerification(ctx fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("resend verification bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	if handler.resendGuard != nil {
		if err := handler.resendGuard.CheckAndConsume(ctx.Context(), req.Email, ctx.IP()); err != nil {
			var limitErr *ResendVerificationRateLimitError
			if errors.As(err, &limitErr) {
				retryAfterSeconds := int64(limitErr.RetryAfter.Seconds())
				if retryAfterSeconds <= 0 {
					retryAfterSeconds = 1
				}
				ctx.Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
				return runtimeError.Respond(
					ctx,
					fiber.StatusTooManyRequests,
					runtimeError.CodeRateLimited,
					"Too many verification requests. Try again later",
				)
			}
			slog.Error("resend verification throttle", "error", err)
			return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Could not resend verification")
		}
	}

	if _, err := handler.service.ResendVerification(req.Email); err != nil {
		if errors.Is(err, ErrEmailAlreadyVerified) {
			return runtimeError.Respond(ctx, fiber.StatusConflict, runtimeError.CodeEmailAlreadyVerified, "Email is already verified")
		}
		if errors.Is(err, user.ErrNotFound) {
			return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
		}
		slog.Error("resend verification", "email", req.Email, "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Could not resend verification")
	}

	slog.Info("verification email sent", "email", req.Email)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
}

// En: ForgotPassword requests a password reset email for the given address.
// Es: ForgotPassword solicita un correo de recuperacion de contraseña para la direccion dada.
func (handler *Handler) ForgotPassword(ctx fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("forgot password bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	if handler.forgotGuard != nil {
		if err := handler.forgotGuard.CheckAndConsume(ctx.Context(), req.Email, ctx.IP()); err != nil {
			var limitErr *ForgotPasswordRateLimitError
			if errors.As(err, &limitErr) {
				retryAfterSeconds := int64(limitErr.RetryAfter.Seconds())
				if retryAfterSeconds <= 0 {
					retryAfterSeconds = 1
				}
				ctx.Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
				return runtimeError.Respond(
					ctx,
					fiber.StatusTooManyRequests,
					runtimeError.CodeRateLimited,
					"Too many password reset requests. Try again later",
				)
			}
			slog.Error("forgot password throttle", "error", err)
			return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Could not process password reset request")
		}
	}

	if err := handler.service.ForgotPassword(ctx.Context(), req.Email); err != nil {
		slog.Error("forgot password", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Could not process password reset request")
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "If the email exists and is eligible, a password reset link has been sent",
	})
}

// En: ResetPassword sets a new password using a one-time reset token from the email link.
// Es: ResetPassword establece una nueva contraseña usando el token de un solo uso del correo.
func (handler *Handler) ResetPassword(ctx fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("reset password bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	if err := handler.service.ResetPassword(req.Token, req.Password); err != nil {
		if errors.Is(err, ErrInvalidPasswordResetToken) {
			return runtimeError.Respond(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeInvalidPasswordResetToken,
				"Invalid or expired password reset token",
			)
		}
		slog.Error("reset password", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Password reset failed")
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Password has been reset successfully"})
}

// En: DevGetVerificationToken returns the current email verification token for a given email.
//
//	This endpoint is intended for development environments only.
//
// Es: DevGetVerificationToken devuelve el token de verificación de correo electrónico actual para un correo dado.
//
//	Este endpoint está pensado solo para entornos de desarrollo.
func (handler *Handler) DevGetVerificationToken(ctx fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("dev get verification token bind error", "error", err)
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(ctx, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	token, err := handler.service.ResendVerification(req.Email)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyVerified) {
			return runtimeError.Respond(ctx, fiber.StatusConflict, runtimeError.CodeEmailAlreadyVerified, "Email is already verified")
		}
		if errors.Is(err, user.ErrNotFound) {
			return runtimeError.Respond(ctx, fiber.StatusNotFound, runtimeError.CodeUserNotFound, "User not found")
		}
		slog.Error("dev get verification token", "email", req.Email, "error", err)
		return runtimeError.Respond(ctx, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Could not generate verification token")
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"token": token,
		},
	})
}

// En: toErrorDetails converts validator.ValidationErrors to runtimeError.ErrorDetail slice.
// Es: toErrorDetails convierte validator.ValidationErrors en un slice de runtimeError.ErrorDetail.
func toErrorDetails(validationErrors validator.ValidationErrors) []runtimeError.ErrorDetail {
	details := make([]runtimeError.ErrorDetail, len(validationErrors))
	for i, fieldError := range validationErrors {
		details[i] = runtimeError.ErrorDetail{
			Field:   fieldError.Field,
			Message: fieldError.Message,
		}
	}
	return details
}
