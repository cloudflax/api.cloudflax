package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	fibercors "github.com/gofiber/fiber/v3/middleware/cors"
)

// CORS returns a Fiber middleware that configures CORS for the API.
// It allows the given origin (typically the frontend URL) to access the API.
func CORS(allowedOrigin string) fiber.Handler {
	origin := strings.TrimSuffix(strings.TrimSpace(allowedOrigin), "/")

	cfg := fibercors.Config{
		AllowOrigins: []string{},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
	}

	// Fail-closed: if no explicit origin is configured, allow none.
	// This prevents accidental open CORS in production.
	if origin != "" {
		cfg.AllowOrigins = []string{origin}
	}

	return fibercors.New(cfg)
}

