package account

import (
	"errors"
	"log/slog"

	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
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
		return runtimeerror.Respond(c, fiber.StatusUnauthorized, runtimeerror.CodeUnauthorized, "Unauthorized")
	}

	var req CreateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("create account bind error", "error", err)
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("create account validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeerror.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeerror.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeValidationError, err.Error())
	}

	account, _, err := h.service.CreateAccount(req.Name, req.Slug, rctx.UserID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserEmailNotVerified):
			return runtimeerror.Respond(c, fiber.StatusForbidden, runtimeerror.CodeEmailVerificationRequired, "Email verification required")
		case errors.Is(err, ErrSlugTaken):
			return runtimeerror.Respond(c, fiber.StatusConflict, runtimeerror.CodeAccountSlugTaken, "Slug is already taken")
		case errors.Is(err, user.ErrNotFound):
			return runtimeerror.Respond(c, fiber.StatusNotFound, runtimeerror.CodeUserNotFound, "User not found")
		default:
			slog.Error("create account", "user_id", rctx.UserID, "error", err)
			return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Failed to create account")
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": account})
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
