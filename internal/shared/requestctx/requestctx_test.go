package requestctx

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// respondWithContext writes the extracted context fields as JSON so tests can assert on them.
func respondWithContext(c fiber.Ctx) error {
	rctx, err := FromFiber(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"user_id":    rctx.UserID,
		"email":      rctx.Email,
		"account_id": rctx.AccountID,
	})
}

func respondWithUserOnly(c fiber.Ctx) error {
	rctx, err := UserOnly(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"user_id":    rctx.UserID,
		"email":      rctx.Email,
		"account_id": rctx.AccountID,
	})
}

func injectLocals(locals map[string]any) fiber.Handler {
	return func(c fiber.Ctx) error {
		for key, value := range locals {
			c.Locals(key, value)
		}
		return c.Next()
	}
}

func TestFromFiber_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test",
		injectLocals(map[string]any{"userID": "user-123", "email": "user@example.com", "accountID": "acc-456"}),
		respondWithContext,
	)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "user-123", result["user_id"])
	assert.Equal(t, "user@example.com", result["email"])
	assert.Equal(t, "acc-456", result["account_id"])
}

func TestFromFiber_MissingUserID(t *testing.T) {
	app := fiber.New()
	app.Get("/test",
		injectLocals(map[string]any{"accountID": "acc-456"}),
		respondWithContext,
	)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestFromFiber_MissingAccountID(t *testing.T) {
	app := fiber.New()
	app.Get("/test",
		injectLocals(map[string]any{"userID": "user-123"}),
		respondWithContext,
	)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestFromFiber_EmailIsOptional(t *testing.T) {
	app := fiber.New()
	app.Get("/test",
		injectLocals(map[string]any{"userID": "user-123", "accountID": "acc-456"}),
		respondWithContext,
	)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "user-123", result["user_id"])
	assert.Empty(t, result["email"])
}

func TestUserOnly_Success(t *testing.T) {
	app := fiber.New()
	app.Get("/test",
		injectLocals(map[string]any{"userID": "user-123", "email": "user@example.com"}),
		respondWithUserOnly,
	)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "user-123", result["user_id"])
	assert.Equal(t, "user@example.com", result["email"])
	assert.Empty(t, result["account_id"])
}

func TestUserOnly_MissingUserID(t *testing.T) {
	app := fiber.New()
	app.Get("/test", injectLocals(map[string]any{}), respondWithUserOnly)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestUserOnly_DoesNotRequireAccountID(t *testing.T) {
	app := fiber.New()
	app.Get("/test",
		injectLocals(map[string]any{"userID": "user-123"}),
		respondWithUserOnly,
	)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
