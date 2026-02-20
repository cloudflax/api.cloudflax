package server

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	require.NoError(t, database.InitForTesting())

	app := fiber.New()
	app.Get("/health", Health())

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"status":"healthy"`)
	assert.Contains(t, string(body), `"db":"connected"`)
}

func TestHome(t *testing.T) {
	app := fiber.New()
	app.Get("/", Home)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"status":"ok"`)
	assert.Contains(t, string(body), `"message":"Cloudflax API"`)
}
