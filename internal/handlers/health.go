package handlers

import (
	"github.com/gofiber/fiber/v3"
)

// Health responde el estado de salud de la API.
func Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "healthy"})
}
