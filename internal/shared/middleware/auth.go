package middleware

import (
	"strings"

	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/gofiber/fiber/v3"
)

// TokenValidator is the interface the auth middleware needs to validate access tokens.
type TokenValidator interface {
	ValidateAccessToken(tokenString string) (userID, email string, err error)
}

// RequireAuth returns a Fiber middleware that validates the Bearer JWT in the
// Authorization header. On success it sets "userID" and "email" in Fiber locals.
func RequireAuth(validator TokenValidator) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Authorization header required")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeTokenInvalid, "Invalid authorization format, expected: Bearer <token>")
		}

		userID, email, err := validator.ValidateAccessToken(parts[1])
		if err != nil {
			return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeTokenInvalid, "Invalid or expired token")
		}

		c.Locals("userID", userID)
		c.Locals("email", email)
		return c.Next()
	}
}
