package auth

import (
	"errors"
	"log/slog"

	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
	"github.com/cloudflax/api.cloudflax/internal/validator"
	"github.com/gofiber/fiber/v3"
)

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
