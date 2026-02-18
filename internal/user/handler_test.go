package user

import (
	"bytes"
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

func TestGetUser_NotFound(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users/:id", handler.GetUser)

	req := httptest.NewRequest("GET", "/users/00000000-0000-0000-0000-000000000000", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeUserNotFound, errResp.Error.Code)
	assert.Equal(t, fiber.StatusNotFound, errResp.Error.Status)
	assert.NotEmpty(t, errResp.Error.Message)
}

func TestGetUser_InvalidUUID(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Get("/users/:id", handler.GetUser)

	// Malformed UUID (extra character in last segment) - previously returned 500
	req := httptest.NewRequest("GET", "/users/a5a10f19-99e6-4ff1-bea4-b2de85f77d061", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeUserNotFound, errResp.Error.Code)
}

func TestGetUser_Found(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Jane", Email: "jane@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Get("/users/:id", handler.GetUser)

	req := httptest.NewRequest("GET", "/users/"+testUser.ID, nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, testUser.ID, result.Data.ID)
	assert.Equal(t, "Jane", result.Data.Name)
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

func TestUpdateUser_Success(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Old Name", Email: "old@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Put("/users/:id", handler.UpdateUser)

	newName := "New Name"
	bodyBytes, _ := json.Marshal(map[string]string{"name": newName})
	req := httptest.NewRequest("PUT", "/users/"+testUser.ID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, newName, result.Data.Name)
	assert.Equal(t, testUser.Email, result.Data.Email)
}

func TestUpdateUser_EmailIgnored(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Original", Email: "original@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Put("/users/:id", handler.UpdateUser)

	// Sending email in body should be ignored; only name and password are updatable.
	body := strings.NewReader(`{"name":"Updated Name","email":"ignored@example.com"}`)
	req := httptest.NewRequest("PUT", "/users/"+testUser.ID, body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "Updated Name", result.Data.Name)
	assert.Equal(t, "original@example.com", result.Data.Email, "email must not be updated")
}

func TestUpdateUser_NotFound(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Put("/users/:id", handler.UpdateUser)

	body := strings.NewReader(`{"name":"Updated"}`)
	req := httptest.NewRequest("PUT", "/users/00000000-0000-0000-0000-000000000000", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeUserNotFound, errResp.Error.Code)
	assert.Equal(t, fiber.StatusNotFound, errResp.Error.Status)
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
