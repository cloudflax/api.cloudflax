package auth

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: TestRoutesDevVerifyEmailTokenMountedOnlyInLocalstack verifies that the dev route
//     is only available when APP_ENV=localstack.
// Es: TestRoutesDevVerifyEmailTokenMountedOnlyInLocalstack verifica que la ruta de
//     desarrollo solo esté disponible cuando APP_ENV=localstack.
func TestRoutesDevVerifyEmailTokenMountedOnlyInLocalstack(test *testing.T) {
	test.Setenv("APP_ENV", "localstack")

	app := fiber.New()
	handler, _ := SetupAuthHandlerTest(test)

	Routes(app, handler, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest("POST", "/auth/dev/verify-email-token", strings.NewReader(`{"email":"dev@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	assert.NotEqual(test, fiber.StatusNotFound, resp.StatusCode, "route must be reachable when APP_ENV=localstack")

	test.Setenv("APP_ENV", "production")

	app2 := fiber.New()
	handler2, _ := SetupAuthHandlerTest(test)

	Routes(app2, handler2, func(c fiber.Ctx) error { return c.Next() })

	req2 := httptest.NewRequest("POST", "/auth/dev/verify-email-token", strings.NewReader(`{"email":"dev@example.com"}`))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app2.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	assert.Equal(test, fiber.StatusNotFound, resp2.StatusCode, "route must not exist when APP_ENV!=localstack")
}

