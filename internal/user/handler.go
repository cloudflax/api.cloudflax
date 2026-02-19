package user

import (
	"errors"
	"log/slog"

	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
	"github.com/cloudflax/api.cloudflax/internal/validator"
	"github.com/gofiber/fiber/v3"
)

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// UpdateMeRequest is the request body for updating the authenticated user's profile.
// At least one field must be present. Email cannot be changed.
type UpdateMeRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=2,max=100"`
	Password *string `json:"password" validate:"omitempty,min=8,max=72"`
}

// Handler handles HTTP requests for users.
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe returns the authenticated user based on the userID stored in locals.
func (h *Handler) GetMe(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized")
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierrors.Respond(c, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found")
		}
		slog.Error("get me", "user_id", userID, "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to get user")
	}
	return c.JSON(fiber.Map{"data": user})
}

// UpdateMe updates the authenticated user's own profile.
func (h *Handler) UpdateMe(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized")
	}

	var req UpdateMeRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("update me bind error", "error", err)
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}
	if req.Name == nil && req.Password == nil {
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, "At least one field (name, password) is required")
	}
	if err := validator.Validate(req); err != nil {
		slog.Debug("update me validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error())
	}

	user, err := h.service.UpdateUser(userID, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierrors.Respond(c, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found")
		}
		slog.Error("update me", "user_id", userID, "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to update user")
	}
	return c.JSON(fiber.Map{"data": user})
}

// CreateUser creates a new user.
func (h *Handler) CreateUser(c fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("create user bind error", "error", err)
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("create user validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return apierrors.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, apierrors.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return apierrors.Respond(c, fiber.StatusBadRequest, apierrors.CodeValidationError, err.Error())
	}

	user, err := h.service.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return apierrors.Respond(c, fiber.StatusConflict, apierrors.CodeEmailAlreadyExists, "Email already exists")
		}
		slog.Error("create user", "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to create user")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": user})
}

// DeleteMe deletes the authenticated user based on the userID stored in locals.
func (h *Handler) DeleteMe(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Unauthorized")
	}

	if err := h.service.DeleteUser(userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return apierrors.Respond(c, fiber.StatusNotFound, apierrors.CodeUserNotFound, "User not found")
		}
		slog.Error("delete me", "user_id", userID, "error", err)
		return apierrors.Respond(c, fiber.StatusInternalServerError, apierrors.CodeInternalServerError, "Failed to delete user")
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
