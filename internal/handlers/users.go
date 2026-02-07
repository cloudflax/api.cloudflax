package handlers

import (
	"log/slog"

	"github.com/cloudflax/api.cloudflax/internal/db"
	"github.com/cloudflax/api.cloudflax/internal/models"
	"github.com/gofiber/fiber/v3"
)

// ListUsers lista todos los usuarios.
func ListUsers(c fiber.Ctx) error {
	var users []models.User
	if err := db.DB.Find(&users).Error; err != nil {
		slog.Error("list users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"users": users,
	})
}

// GetUser obtiene un usuario por ID con sus posts.
func GetUser(c fiber.Ctx) error {
	var user models.User
	if err := db.DB.Preload("Posts").First(&user, "id = ?", c.Params("id")).Error; err != nil {
		slog.Debug("get user not found", "id", c.Params("id"), "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "usuario no encontrado",
		})
	}
	return c.JSON(user)
}
