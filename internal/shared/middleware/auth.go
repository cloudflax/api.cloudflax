package middleware

import (
	"strings"

	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
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
			return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeUnauthorized, "Authorization header required")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeTokenInvalid, "Invalid authorization format, expected: Bearer <token>")
		}

		userID, email, err := validator.ValidateAccessToken(parts[1])
		if err != nil {
			return apierrors.Respond(c, fiber.StatusUnauthorized, apierrors.CodeTokenInvalid, "Invalid or expired token")
		}

		c.Locals("userID", userID)
		c.Locals("email", email)
		return c.Next()
	}
}
