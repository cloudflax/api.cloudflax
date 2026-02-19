package app

import (
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/server"
	"github.com/cloudflax/api.cloudflax/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// Run starts the Fiber server with the loaded configuration.
func Run(cfg *config.Config) error {
	app := fiber.New()

	app.Use(middleware.Logger())
	server.Mount(app, cfg)

	return app.Listen(":" + cfg.Port)
}
