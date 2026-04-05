package auth

import (
	"os"

	"github.com/gofiber/fiber/v3"
)

// En: Routes mounts auth routes on the given router.
// Es: Monta las rutas de autenticación en el router dado.
func Routes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	auth := router.Group("/auth")
	auth.Post("/register", handler.Register)
	auth.Get("/verify-email", handler.VerifyEmail)
	auth.Post("/resend-verification", handler.ResendVerification)
	auth.Post("/forgot-password", handler.ForgotPassword)
	auth.Post("/reset-password", handler.ResetPassword)
	auth.Post("/login", handler.Login)
	auth.Post("/refresh", handler.Refresh)
	auth.Post("/logout", authMiddleware, handler.Logout)

	// Development-only helpers.
	// Mounted in non-production environments (e.g. development or test).
	if os.Getenv("APP_ENV") != "production" {
		auth.Post("/dev/verify-email-token", handler.DevGetVerificationToken)
	}
}
