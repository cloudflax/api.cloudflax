package auth

import (
	"errors"
	"log/slog"

	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/shared/validator"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
)

// RegisterRequest is the request body for POST /auth/register.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// VerifyEmailRequest holds the token from GET /auth/verify-email?token=...
type VerifyEmailRequest struct {
	Token string `query:"token" validate:"required"`
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
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("login validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	pair, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeInvalidCredentials, "Invalid email or password")
		}
		if errors.Is(err, ErrEmailNotVerified) {
			return runtimeError.Respond(c, fiber.StatusForbidden, runtimeError.CodeEmailVerificationRequired, "Email must be verified before you can log in")
		}
		slog.Error("login", "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Login failed")
	}

	return c.JSON(fiber.Map{"data": pair})
}

// Refresh exchanges a valid refresh token for a new token pair.
func (h *Handler) Refresh(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("refresh bind error", "error", err)
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, "refresh_token is required")
	}

	pair, err := h.service.RefreshTokens(req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeTokenInvalid, "Invalid or expired refresh token")
		}
		if errors.Is(err, ErrEmailNotVerified) {
			return runtimeError.Respond(c, fiber.StatusForbidden, runtimeError.CodeEmailVerificationRequired, "Email must be verified before you can use this session")
		}
		slog.Error("refresh tokens", "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Token refresh failed")
	}

	return c.JSON(fiber.Map{"data": pair})
}

// Logout revokes all active refresh tokens for the authenticated user.
func (h *Handler) Logout(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Unauthorized")
	}

	if err := h.service.Logout(userID); err != nil {
		slog.Error("logout", "user_id", userID, "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Logout failed")
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Register creates a new user account with email/password credentials.
// Responds 201 with the created user. A verification email is sent to the
// registered address; the account is not usable until the email is verified.
func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("register bind error", "error", err)
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("register validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	u, _, err := h.service.Register(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrDuplicateEmail) {
			return runtimeError.Respond(c, fiber.StatusConflict, runtimeError.CodeEmailAlreadyExists, "Email already exists")
		}
		slog.Error("register", "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Registration failed")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": u,
		"meta": fiber.Map{"email_verification_required": true},
	})
}

// VerifyEmail marks the user's email as verified using the token from the verification link.
// The token is read from the query string (?token=...) so the link in the email works when clicked (GET).
func (h *Handler) VerifyEmail(c fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := c.Bind().Query(&req); err != nil {
		slog.Debug("verify email bind error", "error", err)
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request")
	}
	if err := validator.Validate(req); err != nil {
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, "token is required")
	}

	if err := h.service.VerifyEmail(req.Token); err != nil {
		if errors.Is(err, ErrInvalidVerificationToken) {
			return runtimeError.Respond(c, fiber.StatusUnprocessableEntity, runtimeError.CodeInvalidVerificationToken, "Invalid or expired verification token")
		}
		slog.Error("verify email", "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Email verification failed")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Email verified successfully"})
}

// ResendVerification generates a new verification token for the given email and
// sends a new verification email via SES.
func (h *Handler) ResendVerification(c fiber.Ctx) error {
	var req ResendVerificationRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("resend verification bind error", "error", err)
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}
	if err := validator.Validate(req); err != nil {
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, "email is required")
	}

	if _, err := h.service.ResendVerification(req.Email); err != nil {
		if errors.Is(err, ErrEmailAlreadyVerified) {
			return runtimeError.Respond(c, fiber.StatusConflict, runtimeError.CodeEmailAlreadyVerified, "Email is already verified")
		}
		if errors.Is(err, user.ErrNotFound) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
		}
		slog.Error("resend verification", "email", req.Email, "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Could not resend verification")
	}

	slog.Info("verification email sent", "email", req.Email)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If the email exists, a verification link has been sent"})
}

// toErrorDetails converts validator.ValidationErrors to runtimeError.ErrorDetail slice.
func toErrorDetails(ve validator.ValidationErrors) []runtimeError.ErrorDetail {
	details := make([]runtimeError.ErrorDetail, len(ve))
	for i, fe := range ve {
		details[i] = runtimeError.ErrorDetail{
			Field:   fe.Field,
			Message: fe.Message,
		}
	}
	return details
}
