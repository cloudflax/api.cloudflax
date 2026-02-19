package user

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts user routes on the given router.
// POST /users (registration) is public; all other endpoints require authentication.
func Routes(router fiber.Router, h *Handler, authMiddleware fiber.Handler) {
	router.Post("/users", h.CreateUser)
	router.Get("/users", authMiddleware, h.ListUser)
	router.Get("/users/me", authMiddleware, h.GetMe)
	router.Delete("/users/:id", authMiddleware, h.DeleteUser)
}
