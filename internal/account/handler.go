package account

import (
	"errors"
	"log/slog"

	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/shared/requestctx"
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
	rctx, err := requestctx.UserOnly(c)
	if err != nil {
		return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Unauthorized")
	}

	var req CreateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("create account bind error", "error", err)
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("create account validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	account, _, err := h.service.CreateAccount(req.Name, req.Slug, rctx.UserID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserEmailNotVerified):
			return runtimeError.Respond(c, fiber.StatusForbidden, runtimeError.CodeEmailVerificationRequired, "Email verification required")
		case errors.Is(err, ErrSlugTaken):
			return runtimeError.Respond(c, fiber.StatusConflict, runtimeError.CodeAccountSlugTaken, "Slug is already taken")
		case errors.Is(err, user.ErrNotFound):
			return runtimeError.Respond(c, fiber.StatusNotFound, runtimeError.CodeUserNotFound, "User not found")
		default:
			slog.Error("create account", "user_id", rctx.UserID, "error", err)
			return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Failed to create account")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": account})
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
