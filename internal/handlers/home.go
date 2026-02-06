package handlers

import (
	"github.com/gofiber/fiber/v3"
)

// Home responde en la ruta raíz.
func Home(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Cloudflax API",
	})
}
