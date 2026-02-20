package auth

import (
	"errors"
	"log/slog"

	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/cloudflax/api.cloudflax/internal/validator"
	"github.com/gofiber/fiber/v3"
)

// RegisterRequest is the request body for POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// VerifyEmailRequest is the request body for POST /auth/verify-email.
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// ResendVerificationRequest is the request body for POST /auth/resend-verification.
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// LoginRequest is the request body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RefreshRequest is the request body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Handler handles HTTP requests for authentication.
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login authenticates a user and returns an access + refresh token pair.
func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("login bind error", "error", err)
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("login validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeerror.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeerror.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeValidationError, err.Error())
	}

	pair, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return runtimeerror.Respond(c, fiber.StatusUnauthorized, runtimeerror.CodeInvalidCredentials, "Invalid email or password")
		}
		slog.Error("login", "error", err)
		return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Login failed")
	}

	return c.JSON(fiber.Map{"data": pair})
}

// Refresh exchanges a valid refresh token for a new token pair.
func (h *Handler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("refresh bind error", "error", err)
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeValidationError, "refresh_token is required")
	}

	pair, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return runtimeerror.Respond(c, fiber.StatusUnauthorized, runtimeerror.CodeTokenInvalid, "Invalid or expired refresh token")
		}
		slog.Error("refresh tokens", "error", err)
		return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Token refresh failed")
	}

	return c.JSON(fiber.Map{"data": pair})
}

// Logout revokes all active refresh tokens for the authenticated user.
func (h *Handler) Logout(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return runtimeerror.Respond(c, fiber.StatusUnauthorized, runtimeerror.CodeUnauthorized, "Unauthorized")
	}

	if err := h.service.Logout(userID); err != nil {
		slog.Error("logout", "user_id", userID, "error", err)
		return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Logout failed")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Register creates a new user account with email/password credentials.
// Responds 201 with the created user. Email verification is required before
// full access; a verification token is generated (and in production would be emailed).
func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("register bind error", "error", err)
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("register validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeerror.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeerror.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeValidationError, err.Error())
	}

	u, _, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			return runtimeerror.Respond(c, fiber.StatusConflict, runtimeerror.CodeEmailAlreadyExists, "Email already exists")
		}
		slog.Error("register", "error", err)
		return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Registration failed")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": u,
		"meta": fiber.Map{"email_verification_required": true},
	})
}

// VerifyEmail marks the user's email as verified using the token from the verification link.
func (h *Handler) VerifyEmail(c fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("verify email bind error", "error", err)
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeValidationError, "token is required")
	}

	if err := h.service.VerifyEmail(req.Token); err != nil {
		if errors.Is(err, ErrInvalidVerificationToken) {
			return runtimeerror.Respond(c, fiber.StatusUnprocessableEntity, runtimeerror.CodeInvalidVerificationToken, "Invalid or expired verification token")
		}
		slog.Error("verify email", "error", err)
		return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Email verification failed")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Email verified successfully"})
}

// ResendVerification generates a new verification token for the given email.
// In production this would send an email; here the token is only logged.
func (h *Handler) ResendVerification(c fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("resend verification bind error", "error", err)
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeValidationError, "email is required")
	}

	token, err := h.service.ResendVerification(req.Email)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyVerified) {
			return runtimeerror.Respond(c, fiber.StatusConflict, runtimeerror.CodeEmailAlreadyVerified, "Email is already verified")
		}
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
		}
		slog.Error("resend verification", "email", req.Email, "error", err)
		return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Could not resend verification")
	}

	slog.Info("email verification token generated", "email", req.Email, "token", token)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
}

// toErrorDetails converts validator.ValidationErrors to runtimeerror.ErrorDetail slice.
func toErrorDetails(ve validator.ValidationErrors) []runtimeerror.ErrorDetail {
	details := make([]runtimeerror.ErrorDetail, len(ve))
	for i, fe := range ve {
		details[i] = runtimeerror.ErrorDetail{
			Field:   fe.Field,
			Message: fe.Message,
		}
	}
	return details
}
