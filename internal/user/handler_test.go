package user

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserHandlerTest(t *testing.T) *Handler {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&User{}))
	repository := NewRepository(database.DB)
	service := NewService(repository)
	return NewHandler(service)
}

// decodeErrorResponse reads the body and decodes it into an apierrors.ErrorResponse.
func decodeErrorResponse(t *testing.T, body io.Reader) apierrors.ErrorResponse {
	t.Helper()
	var result apierrors.ErrorResponse
	require.NoError(t, json.NewDecoder(body).Decode(&result))
	return result
}

func TestGetMe_Success(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Me User", Email: "me@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Get("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", testUser.ID)
		return c.Next()
	}, handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, testUser.ID, result.Data.ID)
	assert.Equal(t, "Me User", result.Data.Name)
	assert.Empty(t, result.Data.PasswordHash)
}

func TestGetMe_Unauthorized(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users/me", handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeUnauthorized, errResp.Error.Code)
}

func TestGetMe_UserNotFound(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", "00000000-0000-0000-0000-000000000000")
		return c.Next()
	}, handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeUserNotFound, errResp.Error.Code)
}

func TestListUser_Empty(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users", handler.ListUser)

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
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Test User", Email: "test@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Get("/users", handler.ListUser)

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

func TestCreateUser_Success(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Data.ID)
	assert.Equal(t, "Alice", result.Data.Name)
	assert.Equal(t, "alice@example.com", result.Data.Email)
	assert.Empty(t, result.Data.PasswordHash)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	handler := setupUserHandlerTest(t)

	existingUser := User{Name: "Existing", Email: "exists@example.com"}
	require.NoError(t, existingUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&existingUser).Error)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"New User","email":"exists@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeEmailAlreadyExists, errResp.Error.Code)
	assert.Equal(t, fiber.StatusConflict, errResp.Error.Status)
}

func TestCreateUser_DuplicateEmail_DifferentName(t *testing.T) {
	// Uniqueness is by email only; different name with same email must be rejected.
	handler := setupUserHandlerTest(t)

	existingUser := User{Name: "Jose Guerrero", Email: "jose.guerrero@cloudflax.com"}
	require.NoError(t, existingUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&existingUser).Error)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"José Guerrero","email":"jose.guerrero@cloudflax.com","password":"123456789"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode, "same email with different name must be rejected")
}

func TestCreateUser_DuplicateEmail_CaseInsensitive(t *testing.T) {
	// Email comparison must be case-insensitive.
	handler := setupUserHandlerTest(t)

	existingUser := User{Name: "Alice", Email: "alice@example.com"}
	require.NoError(t, existingUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&existingUser).Error)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"Bob","email":"Alice@Example.com","password":"password456"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode, "email uniqueness must be case-insensitive")
}

func TestCreateUser_ValidationError_SingleField(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	// Only password is invalid (too short); name and email are valid.
	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"short"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeValidationError, errResp.Error.Code)
	assert.Equal(t, fiber.StatusUnprocessableEntity, errResp.Error.Status)
	require.Len(t, errResp.Error.Details, 1)
	assert.Equal(t, "password", errResp.Error.Details[0].Field)
	assert.NotEmpty(t, errResp.Error.Details[0].Message)
}

func TestCreateUser_ValidationError_MultipleFields(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	// name is empty, email is invalid, password is too short — three field errors.
	body := strings.NewReader(`{"name":"","email":"not-an-email","password":"short"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeValidationError, errResp.Error.Code)
	assert.Equal(t, fiber.StatusUnprocessableEntity, errResp.Error.Status)
	assert.GreaterOrEqual(t, len(errResp.Error.Details), 2, "expected at least 2 field errors")

	fields := make(map[string]string, len(errResp.Error.Details))
	for _, d := range errResp.Error.Details {
		fields[d.Field] = d.Message
	}
	assert.Contains(t, fields, "email")
	assert.Contains(t, fields, "password")
}

func TestDeleteUser_Success(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "To Delete", Email: "delete@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Delete("/users/:id", handler.DeleteUser)

	req := httptest.NewRequest("DELETE", "/users/"+testUser.ID, nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestDeleteUser_NotFound(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Delete("/users/:id", handler.DeleteUser)

	req := httptest.NewRequest("DELETE", "/users/00000000-0000-0000-0000-000000000000", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeUserNotFound, errResp.Error.Code)
}
