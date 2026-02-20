package user

import (
	"errors"
	"log/slog"

	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
	"github.com/cloudflax/api.cloudflax/internal/shared/requestctx"
	"github.com/cloudflax/api.cloudflax/internal/validator"
	"github.com/gofiber/fiber/v3"
)

// Handler handles HTTP requests for users.
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe returns the authenticated user based on the userID stored in locals.
func (handler *Handler) GetMe(ctx fiber.Ctx) error {
	requestContext, err := requestctx.UserOnly(ctx)
	if err != nil {
		return apierrors.Respond(
			ctx, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized",
		)
	}

	user, err := handler.service.GetUser(requestContext.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierrors.Respond(
				ctx, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found",
			)
		}
		slog.Error("get me", "user_id", requestContext.UserID, "error", err)
		return apierrors.Respond(
			ctx, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to get user",
		)
	}

	return ctx.JSON(fiber.Map{"data": user})
}

// UpdateMe updates the authenticated user's own profile.
func (handler *Handler) UpdateMe(ctx fiber.Ctx) error {
	requestContext, err := requestctx.UserOnly(ctx)
	if err != nil {
		return apierrors.Respond(
			ctx, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized",
		)
	}

	var req UpdateMeRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("update me bind error", "error", err)
		return apierrors.Respond(
			ctx, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body",
		)
	}

	if req.Name == nil && req.Password == nil {
		return apierrors.Respond(
			ctx, fiber.StatusBadRequest, apierrors.CodeValidationError, "At least one field (name, password) is required",
		)
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("update me validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(
			ctx, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error(),
		)
	}

	user, err := handler.service.UpdateUser(requestContext.UserID, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierrors.Respond(
				ctx, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found",
			)
		}
		slog.Error("update me", "user_id", requestContext.UserID, "error", err)
		return apierrors.Respond(
			ctx, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to update user",
		)
	}

	return ctx.JSON(fiber.Map{"data": user})
}

// CreateUser creates a new user.
func (handler *Handler) CreateUser(ctx fiber.Ctx) error {
	var req CreateUserRequest
	if err := ctx.Bind().Body(&req); err != nil {
		slog.Debug("create user bind error", "error", err)
		return apierrors.Respond(
			ctx, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body",
		)
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("create user validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				ctx, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(
			ctx, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error(),
		)
	}

	user, err := handler.service.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return apierrors.Respond(
				ctx, fiber.StatusConflict, apierrors.CodeEmailAlreadyExists, "Email already exists",
			)
		}
		slog.Error("create user", "error", err)
		return apierrors.Respond(
			ctx, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to create user",
		)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"data": user})
}

// DeleteMe deletes the authenticated user based on the userID stored in locals.
func (handler *Handler) DeleteMe(ctx fiber.Ctx) error {
	requestContext, err := requestctx.UserOnly(ctx)
	if err != nil {
		return apierrors.Respond(
			ctx, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized",
		)
	}

	if err := handler.service.DeleteUser(requestContext.UserID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierrors.Respond(
				ctx, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found",
			)
		}
		slog.Error("delete me", "user_id", requestContext.UserID, "error", err)
		return apierrors.Respond(
			ctx, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to delete user",
		)
	}

	return ctx.Status(fiber.StatusNoContent).Send(nil)
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
