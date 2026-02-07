package app

import (
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/handlers"
	"github.com/gofiber/fiber/v3"
)

// Run inicia el servidor Fiber con la configuración cargada.
func Run(cfg *config.Config) error {
	app := fiber.New()

	// Rutas
	app.Get("/", handlers.Home)
	app.Get("/health", handlers.Health(cfg))

	return app.Listen(":" + cfg.Port)
}
