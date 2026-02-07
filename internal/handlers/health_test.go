package handlers

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth(t *testing.T) {
	require.NoError(t, db.InitForTesting())

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
