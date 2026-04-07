package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

// SecurityHeaders sets defensive HTTP response headers for browser clients.
// Strict-Transport-Security is not set here; HSTS belongs on the TLS terminator (load balancer).
func SecurityHeaders() fiber.Handler {
	return helmet.New(helmet.Config{
		XSSProtection:             "0",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		ReferrerPolicy:            "no-referrer",
		CrossOriginEmbedderPolicy: "unsafe-none",
		CrossOriginOpenerPolicy:   "unsafe-none",
		CrossOriginResourcePolicy: "cross-origin",
		HSTSMaxAge:                0,
	})
}
