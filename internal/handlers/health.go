package handlers

import (
	"context"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/db"
	"github.com/gofiber/fiber/v3"
)

const defaultRequestTimeout = 5 * time.Second

// Health devuelve un handler que verifica la salud de la API y la conexión a PostgreSQL.
func Health(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
		defer cancel()

		if err := db.Ping(ctx, cfg); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "unhealthy",
				"db":     err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status": "healthy",
			"db":     "connected",
		})
	}
}
