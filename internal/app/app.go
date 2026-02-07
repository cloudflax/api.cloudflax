package app

import (
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/handlers"
	"github.com/cloudflax/api.cloudflax/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

// Run inicia el servidor Fiber con la configuración cargada.
func Run(cfg *config.Config) error {
	app := fiber.New()

	app.Use(middleware.Logger())

	// Rutas
	app.Get("/", handlers.Home)
	app.Get("/health", handlers.Health())
	app.Get("/users", handlers.ListUsers)
	app.Get("/users/:id", handlers.GetUser)

	return app.Listen(":" + cfg.Port)
}
