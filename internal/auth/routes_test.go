package auth

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: TestRoutesDevVerifyEmailTokenMountedWhenFlag verifies the dev route mounts only when the caller passes mountDevVerifyEmailToken.
// Es: TestRoutesDevVerifyEmailTokenMountedWhenFlag verifica el montaje de la ruta dev según el flag explícito.
func TestRoutesDevVerifyEmailTokenMountedWhenFlag(test *testing.T) {
	app := fiber.New()
	handler, authService := SetupAuthHandlerTest(test)

	_, _, err := authService.Register("Dev User", "dev@example.com", "password123")
	require.NoError(test, err)

	Routes(app, handler, func(c fiber.Ctx) error { return c.Next() }, true)

	req := httptest.NewRequest("POST", "/auth/dev/verify-email-token", strings.NewReader(`{"email":"dev@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	assert.NotEqual(test, fiber.StatusNotFound, resp.StatusCode, "route must be reachable when mount flag is true")

	app2 := fiber.New()
	handler2, _ := SetupAuthHandlerTest(test)

	Routes(app2, handler2, func(c fiber.Ctx) error { return c.Next() }, false)

	req2 := httptest.NewRequest("POST", "/auth/dev/verify-email-token", strings.NewReader(`{"email":"dev@example.com"}`))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app2.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	assert.Equal(test, fiber.StatusNotFound, resp2.StatusCode, "route must not exist when mount flag is false")
}
