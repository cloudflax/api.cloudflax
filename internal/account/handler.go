package account

import (
	"errors"
	"log/slog"

	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/cloudflax/api.cloudflax/internal/validator"
	"github.com/gofiber/fiber/v3"
)

// Handler handles HTTP requests for accounts.
type Handler struct {
	service *Service
}

// NewHandler creates a new account handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateAccount handles POST /accounts.
// Creates an account owned by the authenticated user. Requires a verified email.
func (h *Handler) CreateAccount(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized")
	}

	var req CreateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("create account bind error", "error", err)
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("create account validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error())
	}

	account, _, err := h.service.CreateAccount(req.Name, req.Slug, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserEmailNotVerified):
			return apierrors.Respond(c, fiber.StatusForbidden, apierrors.CodeEmailVerificationRequired, "Email verification required")
		case errors.Is(err, ErrSlugTaken):
			return apierrors.Respond(c, fiber.StatusConflict, apierrors.CodeAccountSlugTaken, "Slug is already taken")
		case errors.Is(err, user.ErrNotFound):
			return apierrors.Respond(c, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found")
		default:
			slog.Error("create account", "user_id", userID, "error", err)
			return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to create account")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": account})
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
