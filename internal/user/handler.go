package user

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// CreateUserRequest is the request body for creating a user.
type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserRequest is the request body for updating a user.
// Only name and password can be updated; email is not allowed.
type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Password *string `json:"password"`
}

// Handler handles HTTP requests for users.
type Handler struct {
	service *Service
}

// NewHandler creates a new user handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListUser lists all users.
func (h *Handler) ListUser(c fiber.Ctx) error {
	users, err := h.service.ListUser()
	if err != nil {
		slog.Error("list users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": users,
	})
}

// GetUser gets a user by ID.
func (h *Handler) GetUser(c fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.service.GetUser(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			slog.Debug("get user not found", "id", id, "error", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		slog.Error("get user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": user,
	})
}

// CreateUser creates a new user.
func (h *Handler) CreateUser(c fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("create user bind error", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "name, email and password are required",
		})
	}
	user, err := h.service.CreateUser(req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "email already exists",
			})
		}
		slog.Error("create user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": user,
	})
}

// UpdateUser updates a user by ID.
func (h *Handler) UpdateUser(c fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("update user bind error", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.Name == nil && req.Password == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "at least one field (name, password) is required",
		})
	}
	user, err := h.service.UpdateUser(id, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		slog.Error("update user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"data": user,
	})
}

// DeleteUser deletes a user by ID.
func (h *Handler) DeleteUser(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.DeleteUser(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		slog.Error("delete user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusNoContent).Send(nil)
}
