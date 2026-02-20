package user

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts user routes on the given router.
func Routes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	//router.Post("/users", authMiddleware, handler.CreateUser)
	router.Get("/users/me", authMiddleware, handler.GetMe)
	router.Put("/users/me", authMiddleware, handler.UpdateMe)
	router.Delete("/users/me", authMiddleware, handler.DeleteMe)
}
