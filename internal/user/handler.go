package user

import (
	"errors"
	"log/slog"

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
				"error": "usuario no encontrado",
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
