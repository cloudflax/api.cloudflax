package user

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
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

// En: decodeErrorResponse reads the body and decodes it into a runtimeerror.ErrorResponse.
// Es: decodeErrorResponse lee el cuerpo y lo decodifica en un runtimeerror.ErrorResponse.
func decodeErrorResponse(t *testing.T, body io.Reader) runtimeerror.ErrorResponse {
	t.Helper()
	var result runtimeerror.ErrorResponse
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
	assert.Equal(t, runtimeerror.CodeUnauthorized, errResp.Error.Code)
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
	assert.Equal(t, runtimeerror.CodeUserNotFound, errResp.Error.Code)
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
	assert.Equal(t, runtimeerror.CodeEmailAlreadyExists, errResp.Error.Code)
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
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
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
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
	assert.Equal(t, fiber.StatusUnprocessableEntity, errResp.Error.Status)
	assert.GreaterOrEqual(t, len(errResp.Error.Details), 2, "expected at least 2 field errors")

	fields := make(map[string]string, len(errResp.Error.Details))
	for _, d := range errResp.Error.Details {
		fields[d.Field] = d.Message
	}
	assert.Contains(t, fields, "email")
	assert.Contains(t, fields, "password")
}

func setupDeleteMe(t *testing.T, handler *Handler, userID string) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Delete("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}, handler.DeleteMe)
	return app
}

func TestDeleteMe_Success(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Me To Delete", Email: "deleteme@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := setupDeleteMe(t, handler, testUser.ID)

	req := httptest.NewRequest("DELETE", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	var deleted User
	dbResult := database.DB.Unscoped().First(&deleted, "id = ?", testUser.ID)
	require.NoError(t, dbResult.Error)
	assert.True(t, deleted.DeletedAt.Valid, "user should be soft-deleted")
}

func TestDeleteMe_Unauthorized(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Delete("/users/me", handler.DeleteMe)

	req := httptest.NewRequest("DELETE", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeUnauthorized, errResp.Error.Code)
}

func TestDeleteMe_UserNotFound(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := setupDeleteMe(t, handler, "00000000-0000-0000-0000-000000000000")

	req := httptest.NewRequest("DELETE", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeUserNotFound, errResp.Error.Code)
}

func setupUpdateMe(t *testing.T, handler *Handler, userID string) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Put("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}, handler.UpdateMe)
	return app
}

func TestUpdateMe_Unauthorized(t *testing.T) {
	handler := setupUserHandlerTest(t)

	app := fiber.New()
	app.Put("/users/me", handler.UpdateMe)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeUnauthorized, errResp.Error.Code)
}

func TestUpdateMe_NoFieldsProvided(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Original", Email: "original@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(t, handler, testUser.ID)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
}

func TestUpdateMe_UpdateName(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Old Name", Email: "updateme@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(t, handler, testUser.ID)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"New Name"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "New Name", result.Data.Name)
	assert.Equal(t, testUser.Email, result.Data.Email)
	assert.Empty(t, result.Data.PasswordHash)
}

func TestUpdateMe_EmailIgnored(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Alice", Email: "alice.me@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(t, handler, testUser.ID)

	// email field in body must be silently ignored.
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"Alice Updated","email":"hacker@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "Alice Updated", result.Data.Name)
	assert.Equal(t, "alice.me@example.com", result.Data.Email, "email must not be updated")
}

func TestUpdateMe_ValidationError(t *testing.T) {
	handler := setupUserHandlerTest(t)

	testUser := User{Name: "Bob", Email: "bob.me@example.com"}
	require.NoError(t, testUser.SetPassword("secret123"))
	require.NoError(t, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(t, handler, testUser.ID)

	// name too short (1 char), password too short.
	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"X","password":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
	assert.GreaterOrEqual(t, len(errResp.Error.Details), 1)
}
