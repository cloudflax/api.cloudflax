package user

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
)

// Handler handles HTTP requests for users.
type Handler struct {
	svc *Service
}

// NewHandler creates a new user handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListUser lists all users.
func (h *Handler) ListUser(c fiber.Ctx) error {
	users, err := h.svc.ListUser()
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
	u, err := h.svc.GetUser(id)
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
		"data": u,
	})
}
