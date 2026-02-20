package server

import (
	"context"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/gofiber/fiber/v3"
)

const defaultRequestTimeout = 5 * time.Second

// Health returns a handler that checks API health and PostgreSQL connection.
func Health() fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultRequestTimeout)
		defer cancel()

		if err := database.Ping(ctx); err != nil {
			return runtimeError.Respond(c, fiber.StatusServiceUnavailable, runtimeError.CodeInternalServerError, "Service unavailable")
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
