package server

import (
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
)

// Mount mounts all routes on the Fiber app.
func Mount(app *fiber.App) {
	app.Get("/", Home)
	app.Get("/health", Health())

	userRepository := user.NewRepository(database.DB)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)
	user.Routes(app, userHandler)
}
