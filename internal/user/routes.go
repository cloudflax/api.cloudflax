package user

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts user routes on the given router.
func Routes(router fiber.Router, h *Handler, authMiddleware fiber.Handler) {
	//router.Post("/users", authMiddleware, h.CreateUser)
	router.Get("/users/me", authMiddleware, h.GetMe)
	router.Put("/users/me", authMiddleware, h.UpdateMe)
	router.Delete("/users/me", authMiddleware, h.DeleteMe)
}
