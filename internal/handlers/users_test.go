package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/db"
	"github.com/cloudflax/api.cloudflax/internal/models"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsersTest(t *testing.T) {
	t.Helper()
	require.NoError(t, db.InitForTesting())
}

func TestListUsers_Empty(t *testing.T) {
	setupUsersTest(t)

	app := fiber.New()
	app.Get("/users", ListUsers)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	users, ok := result["users"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, users)
}

func TestListUsers_WithData(t *testing.T) {
	setupUsersTest(t)

	user := models.User{Name: "Test User", Email: "test@example.com"}
	require.NoError(t, user.SetPassword("secret123"))
	require.NoError(t, db.DB.Create(&user).Error)

	app := fiber.New()
	app.Get("/users", ListUsers)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Users []models.User `json:"users"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result.Users, 1)
	assert.Equal(t, "Test User", result.Users[0].Name)
}

func TestGetUser_NotFound(t *testing.T) {
	setupUsersTest(t)

	app := fiber.New()
	app.Get("/users/:id", GetUser)

	req := httptest.NewRequest("GET", "/users/00000000-0000-0000-0000-000000000000", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetUser_Found(t *testing.T) {
	setupUsersTest(t)

	user := models.User{Name: "Jane", Email: "jane@example.com"}
	require.NoError(t, user.SetPassword("secret123"))
	require.NoError(t, db.DB.Create(&user).Error)

	app := fiber.New()
	app.Get("/users/:id", GetUser)

	req := httptest.NewRequest("GET", "/users/"+user.ID, nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result models.User
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, "Jane", result.Name)
}
