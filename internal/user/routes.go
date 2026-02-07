package user

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts user routes on the given router.
func Routes(router fiber.Router, h *Handler) {
	router.Get("/users", h.ListUser)
	router.Get("/users/:id", h.GetUser)
	router.Post("/users", h.CreateUser)
	router.Put("/users/:id", h.UpdateUser)
	router.Delete("/users/:id", h.DeleteUser)
}
