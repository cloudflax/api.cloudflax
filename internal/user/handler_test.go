package user

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserHandlerTest(t *testing.T) *Handler {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&User{}))
	repo := NewRepository(database.DB)
	svc := NewService(repo)
	return NewHandler(svc)
}

func TestListUser_Empty(t *testing.T) {
	h := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users", h.ListUser)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &result))
	users, ok := result["data"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, users)
}

func TestListUser_WithData(t *testing.T) {
	h := setupUserHandlerTest(t)

	u := User{Name: "Test User", Email: "test@example.com"}
	require.NoError(t, u.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&u).Error)

	app := fiber.New()
	app.Get("/users", h.ListUser)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []User `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	require.Len(t, result.Data, 1)
	assert.Equal(t, "Test User", result.Data[0].Name)
}

func TestGetUser_NotFound(t *testing.T) {
	h := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users/:id", h.GetUser)

	req := httptest.NewRequest("GET", "/users/00000000-0000-0000-0000-000000000000", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetUser_Found(t *testing.T) {
	h := setupUserHandlerTest(t)

	u := User{Name: "Jane", Email: "jane@example.com"}
	require.NoError(t, u.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&u).Error)

	app := fiber.New()
	app.Get("/users/:id", h.GetUser)

	req := httptest.NewRequest("GET", "/users/"+u.ID, nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, u.ID, result.Data.ID)
	assert.Equal(t, "Jane", result.Data.Name)
}
