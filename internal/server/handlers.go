package server

import (
	"context"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/gofiber/fiber/v3"
)

const defaultRequestTimeout = 5 * time.Second

// Health returns a handler that checks API health and PostgreSQL connection.
func Health() fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
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

// Home responds on the root path.
func Home(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "Cloudflax API",
	})
}
