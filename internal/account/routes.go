package account

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts account routes on the given router.
func Routes(router fiber.Router, h *Handler, authMiddleware fiber.Handler) {
	router.Post("/accounts", authMiddleware, h.CreateAccount)
	router.Post("/accounts/active", authMiddleware, h.SetActiveAccount)
}
