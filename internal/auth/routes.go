package auth

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts auth routes on the given router.
func Routes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	auth := router.Group("/auth")
	auth.Post("/register", handler.Register)
	auth.Get("/verify-email", handler.VerifyEmail)
	auth.Post("/resend-verification", handler.ResendVerification)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.Refresh)
	auth.Post("/logout", authMiddleware, handler.Logout)
}
