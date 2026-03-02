package auth

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: TestRoutesDevVerifyEmailTokenMountedOnlyInNonProduction verifies that the dev route
// is only available when APP_ENV is not production.
// Es: TestRoutesDevVerifyEmailTokenMountedOnlyInNonProduction verifica que la ruta de
// desarrollo solo esté disponible cuando APP_ENV no es production.
func TestRoutesDevVerifyEmailTokenMountedOnlyInNonProduction(test *testing.T) {
	require.NoError(test, os.Setenv("APP_ENV", "development"))

	app := fiber.New()
	handler, authService := SetupAuthHandlerTest(test)

	_, _, err := authService.Register("Dev User", "dev@example.com", "password123")
	require.NoError(test, err)

	Routes(app, handler, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest("POST", "/auth/dev/verify-email-token", strings.NewReader(`{"email":"dev@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	assert.NotEqual(test, fiber.StatusNotFound, resp.StatusCode, "route must be reachable when APP_ENV is non-production")

	require.NoError(test, os.Setenv("APP_ENV", "production"))

	app2 := fiber.New()
	handler2, _ := SetupAuthHandlerTest(test)

	Routes(app2, handler2, func(c fiber.Ctx) error { return c.Next() })

	req2 := httptest.NewRequest("POST", "/auth/dev/verify-email-token", strings.NewReader(`{"email":"dev@example.com"}`))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app2.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	assert.Equal(test, fiber.StatusNotFound, resp2.StatusCode, "route must not exist when APP_ENV=production")
}
