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

	userRepo := user.NewRepository(database.DB)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)
	user.Routes(app, userHandler)
}
