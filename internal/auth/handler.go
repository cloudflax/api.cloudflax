package auth

import (
	"errors"
	"log/slog"

	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
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
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("login validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error())
	}

	pair, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeInvalidCredentials, "Invalid email or password")
		}
		slog.Error("login", "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Login failed")
	}

	return c.JSON(fiber.Map{"data": pair})
}

// Refresh exchanges a valid refresh token for a new token pair.
func (h *Handler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("refresh bind error", "error", err)
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, "refresh_token is required")
	}

	pair, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeTokenInvalid, "Invalid or expired refresh token")
		}
		slog.Error("refresh tokens", "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Token refresh failed")
	}

	return c.JSON(fiber.Map{"data": pair})
}

// Logout revokes all active refresh tokens for the authenticated user.
func (h *Handler) Logout(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized")
	}

	if err := h.service.Logout(userID); err != nil {
		slog.Error("logout", "user_id", userID, "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Logout failed")
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
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("register validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error())
	}

	u, _, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			return apierrors.Respond(c, fiber.StatusConflict, apierrors.CodeEmailAlreadyExists, "Email already exists")
		}
		slog.Error("register", "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Registration failed")
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
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, "token is required")
	}

	if err := h.service.VerifyEmail(req.Token); err != nil {
		if errors.Is(err, ErrInvalidVerificationToken) {
			return apierrors.Respond(c, fiber.StatusUnprocessableEntity, apierrors.CodeInvalidVerificationToken, "Invalid or expired verification token")
		}
		slog.Error("verify email", "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Email verification failed")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Email verified successfully"})
}

// ResendVerification generates a new verification token for the given email.
// In production this would send an email; here the token is only logged.
func (h *Handler) ResendVerification(c fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("resend verification bind error", "error", err)
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, "email is required")
	}

	token, err := h.service.ResendVerification(req.Email)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyVerified) {
			return apierrors.Respond(c, fiber.StatusConflict, apierrors.CodeEmailAlreadyVerified, "Email is already verified")
		}
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
		}
		slog.Error("resend verification", "email", req.Email, "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Could not resend verification")
	}

	slog.Info("email verification token generated", "email", req.Email, "token", token)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
}

// toErrorDetails converts validator.ValidationErrors to apierrors.ErrorDetail slice.
func toErrorDetails(ve validator.ValidationErrors) []apierrors.ErrorDetail {
	details := make([]apierrors.ErrorDetail, len(ve))
	for i, fe := range ve {
		details[i] = apierrors.ErrorDetail{
			Field:   fe.Field,
			Message: fe.Message,
		}
	}
	return details
}
