package app

import (
	"github.com/cloudflax/api.cloudflax/internal/handlers"
	"github.com/gofiber/fiber/v3"
)

// Run inicia el servidor Fiber en el puerto indicado.
func Run(port string) error {
	app := fiber.New()

	// Rutas
	app.Get("/", handlers.Home)
	app.Get("/health", handlers.Health)

	return app.Listen(":" + port)
}
