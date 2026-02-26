package user

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
)

// SetupUserHandlerTest initializes the test environment and returns a user handler.
// En: Sets up database connection, runs migrations, and creates a new user handler for testing.
// Es: Configura la conexión a la base de datos, ejecuta las migraciones y crea un nuevo manejador de usuarios para pruebas.
func SetupUserHandlerTest(test *testing.T) *Handler {
	test.Helper()
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&User{}))
	repository := NewRepository(database.DB)
	service := NewService(repository)
	return NewHandler(service)
}

type stubAccountLister struct {
	accounts []AccountSummary
	err      error
}

func (s *stubAccountLister) ListAccountsForUser(userID string) ([]AccountSummary, error) {
	return s.accounts, s.err
}

// DecodeErrorResponse reads the response body and decodes it into an ErrorResponse.
// En: Helper function to decode HTTP error responses from the API.
// Es: Función auxiliar para decodificar respuestas de error HTTP de la API.
func DecodeErrorResponse(test *testing.T, body io.Reader) runtimeError.ErrorResponse {
	test.Helper()
	var result runtimeError.ErrorResponse
	require.NoError(test, json.NewDecoder(body).Decode(&result))
	return result
}

// TestGetMeSuccess tests successful retrieval of the authenticated user.
// En: Verifies that a user can successfully get their own profile information.
// Es: Verifica que un usuario pueda obtener exitosamente su propia información de perfil.
func TestGetMeSuccess(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	testUser := User{Name: "Me User", Email: "me@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Get("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", testUser.ID)
		return c.Next()
	}, handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(test, testUser.ID, result.Data.ID)
	assert.Equal(test, "Me User", result.Data.Name)
	assert.Empty(test, result.Data.PasswordHash)
	// GET /users/me is the only endpoint that returns active_account_id
	assert.Nil(test, result.Data.ActiveAccountID, "new user has no active account")
}

// TestGetMeReturnsActiveAccountID ensures GET /users/me includes active_account_id when set.
func TestGetMeReturnsActiveAccountID(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	activeID := "a0b1c2d3-e4f5-6789-abcd-ef0123456789"
	testUser := User{Name: "With Account", Email: "with-account@example.com", ActiveAccountID: &activeID}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Get("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", testUser.ID)
		return c.Next()
	}, handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	require.NotNil(test, result.Data.ActiveAccountID)
	assert.Equal(test, activeID, *result.Data.ActiveAccountID)
}

func TestGetMyAccountsSuccess(test *testing.T) {
	repository := NewRepository(database.DB)
	service := NewService(repository)

	stub := &stubAccountLister{
		accounts: []AccountSummary{
			{ID: "acc-1", Name: "First", Slug: "first"},
			{ID: "acc-2", Name: "Second", Slug: "second"},
		},
	}
	handler := NewHandler(service).WithAccountLister(stub)

	testUser := User{Name: "Me User", Email: "me.accounts@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := fiber.New()
	app.Get("/users/me/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", testUser.ID)
		return c.Next()
	}, handler.GetMyAccounts)

	req := httptest.NewRequest("GET", "/users/me/accounts", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data []AccountSummary `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.Len(test, result.Data, 2)
	assert.Equal(test, "First", result.Data[0].Name)
}

func TestGetMyAccountsUnauthorized(test *testing.T) {
	repository := NewRepository(database.DB)
	service := NewService(repository)
	handler := NewHandler(service).WithAccountLister(&stubAccountLister{})

	app := fiber.New()
	app.Get("/users/me/accounts", handler.GetMyAccounts)

	req := httptest.NewRequest("GET", "/users/me/accounts", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUnauthorized, errorResponse.Error.Code)
}

func TestGetMyAccountsUserNotFound(test *testing.T) {
	repository := NewRepository(database.DB)
	service := NewService(repository)
	stub := &stubAccountLister{
		err: ErrNotFound,
	}
	handler := NewHandler(service).WithAccountLister(stub)

	app := fiber.New()
	app.Get("/users/me/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", "00000000-0000-0000-0000-000000000000")
		return c.Next()
	}, handler.GetMyAccounts)

	req := httptest.NewRequest("GET", "/users/me/accounts", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusNotFound, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUserNotFound, errorResponse.Error.Code)
}

func TestGetMyAccountsInternalError(test *testing.T) {
	repository := NewRepository(database.DB)
	service := NewService(repository)
	stub := &stubAccountLister{
		err: errors.New("boom"),
	}
	handler := NewHandler(service).WithAccountLister(stub)

	app := fiber.New()
	app.Get("/users/me/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", "00000000-0000-0000-0000-000000000000")
		return c.Next()
	}, handler.GetMyAccounts)

	req := httptest.NewRequest("GET", "/users/me/accounts", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusInternalServerError, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeInternalServerError, errorResponse.Error.Code)
}

// TestGetMeUnauthorized tests that unauthenticated requests are rejected.
// En: Ensures that requests without authentication return a 401 Unauthorized error.
// Es: Asegura que las solicitudes sin autenticación devuelvan un error 401 No Autorizado.
func TestGetMeUnauthorized(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Get("/users/me", handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUnauthorized, errorResponse.Error.Code)
}

// TestGetMeUserNotFound tests behavior when authenticated user doesn't exist.
// En: Verifies that a 404 error is returned when the user ID in context doesn't exist.
// Es: Verifica que se devuelva un error 404 cuando el ID de usuario en el contexto no existe.
func TestGetMeUserNotFound(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Get("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", "00000000-0000-0000-0000-000000000000")
		return c.Next()
	}, handler.GetMe)

	req := httptest.NewRequest("GET", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusNotFound, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUserNotFound, errorResponse.Error.Code)
}

// TestCreateUserSuccess tests successful user creation.
// En: Verifies that a new user can be created with valid data.
// Es: Verifica que un nuevo usuario pueda ser creado con datos válidos.
func TestCreateUserSuccess(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(test, result.Data.ID)
	assert.Equal(test, "Alice", result.Data.Name)
	assert.Equal(test, "alice@example.com", result.Data.Email)
	assert.Empty(test, result.Data.PasswordHash)
}

// TestCreateUserDuplicateEmail tests that duplicate email addresses are rejected.
// En: Ensures that creating a user with an existing email returns a conflict error.
// Es: Asegura que crear un usuario con un email existente devuelva un error de conflicto.
func TestCreateUserDuplicateEmail(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	existingUser := User{Name: "Existing", Email: "exists@example.com"}
	require.NoError(test, existingUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&existingUser).Error)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"New User","email":"exists@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusConflict, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeEmailAlreadyExists, errorResponse.Error.Code)
	assert.Equal(test, fiber.StatusConflict, errorResponse.Error.Status)
}

// TestCreateUserDuplicateEmailDifferentName tests email uniqueness regardless of name.
// En: Verifies that email uniqueness is enforced even when the name is different.
// Es: Verifica que la unicidad del email se aplique incluso cuando el nombre es diferente.
func TestCreateUserDuplicateEmailDifferentName(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	existingUser := User{Name: "Jose Guerrero", Email: "jose.guerrero@cloudflax.com"}
	require.NoError(test, existingUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&existingUser).Error)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"José Guerrero","email":"jose.guerrero@cloudflax.com","password":"123456789"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusConflict, resp.StatusCode, "same email with different name must be rejected")
}

// TestCreateUserDuplicateEmailCaseInsensitive tests case-insensitive email uniqueness.
// En: Ensures that email uniqueness check is case-insensitive.
// Es: Asegura que la verificación de unicidad del email no distinga entre mayúsculas y minúsculas.
func TestCreateUserDuplicateEmailCaseInsensitive(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	existingUser := User{Name: "Alice", Email: "alice@example.com"}
	require.NoError(test, existingUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&existingUser).Error)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"Bob","email":"Alice@Example.com","password":"password456"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusConflict, resp.StatusCode, "email uniqueness must be case-insensitive")
}

// TestCreateUserValidationErrorSingleField tests validation for a single invalid field.
// En: Verifies that validation errors are properly returned for a single invalid field.
// Es: Verifica que los errores de validación se devuelvan correctamente para un solo campo inválido.
func TestCreateUserValidationErrorSingleField(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"short"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnprocessableEntity, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeValidationError, errorResponse.Error.Code)
	assert.Equal(test, fiber.StatusUnprocessableEntity, errorResponse.Error.Status)
	require.Len(test, errorResponse.Error.Details, 1)
	assert.Equal(test, "password", errorResponse.Error.Details[0].Field)
	assert.NotEmpty(test, errorResponse.Error.Details[0].Message)
}

// TestCreateUserValidationErrorMultipleFields tests validation for multiple invalid fields.
// En: Verifies that validation errors for multiple fields are properly returned.
// Es: Verifica que los errores de validación para múltiples campos se devuelvan correctamente.
func TestCreateUserValidationErrorMultipleFields(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Post("/users", handler.CreateUser)

	body := strings.NewReader(`{"name":"","email":"not-an-email","password":"short"}`)
	req := httptest.NewRequest("POST", "/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnprocessableEntity, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeValidationError, errorResponse.Error.Code)
	assert.Equal(test, fiber.StatusUnprocessableEntity, errorResponse.Error.Status)
	assert.GreaterOrEqual(test, len(errorResponse.Error.Details), 2, "expected at least 2 field errors")

	fields := make(map[string]string, len(errorResponse.Error.Details))
	for _, detail := range errorResponse.Error.Details {
		fields[detail.Field] = detail.Message
	}
	assert.Contains(test, fields, "email")
	assert.Contains(test, fields, "password")
}

// setupDeleteMe creates a Fiber app with the DeleteMe handler and authentication middleware.
// En: Helper function to set up the test environment for Delete Me endpoint tests.
// Es: Función auxiliar para configurar el entorno de prueba para los endpoints de eliminación.
func setupDeleteMe(test *testing.T, handler *Handler, userID string) *fiber.App {
	test.Helper()
	app := fiber.New()
	app.Delete("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}, handler.DeleteMe)
	return app
}

// TestDeleteMeSuccess tests successful user deletion.
// En: Verifies that an authenticated user can successfully delete their own account.
// Es: Verifica que un usuario autenticado pueda eliminar exitosamente su propia cuenta.
func TestDeleteMeSuccess(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	testUser := User{Name: "Me To Delete", Email: "deleteme@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := setupDeleteMe(test, handler, testUser.ID)

	req := httptest.NewRequest("DELETE", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusNoContent, resp.StatusCode)

	var deleted User
	databaseResult := database.DB.Unscoped().First(&deleted, "id = ?", testUser.ID)
	require.NoError(test, databaseResult.Error)
	assert.True(test, deleted.DeletedAt.Valid, "user should be soft-deleted")
}

// TestDeleteMeUnauthorized tests that unauthenticated delete requests are rejected.
// En: Ensures that delete requests without authentication return a 401 error.
// Es: Asegura que las solicitudes de eliminación sin autenticación devuelvan un error 401.
func TestDeleteMeUnauthorized(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Delete("/users/me", handler.DeleteMe)

	req := httptest.NewRequest("DELETE", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUnauthorized, errorResponse.Error.Code)
}

// TestDeleteMeUserNotFound tests delete behavior when user doesn't exist.
// En: Verifies that a 404 error is returned when trying to delete a non-existent user.
// Es: Verifica que se devuelva un error 404 al intentar eliminar un usuario que no existe.
func TestDeleteMeUserNotFound(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := setupDeleteMe(test, handler, "00000000-0000-0000-0000-000000000000")

	req := httptest.NewRequest("DELETE", "/users/me", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusNotFound, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUserNotFound, errorResponse.Error.Code)
}

// setupUpdateMe creates a Fiber app with the UpdateMe handler and authentication middleware.
// En: Helper function to set up the test environment for Update Me endpoint tests.
// Es: Función auxiliar para configurar el entorno de prueba para los endpoints de actualización.
func setupUpdateMe(test *testing.T, handler *Handler, userID string) *fiber.App {
	test.Helper()
	app := fiber.New()
	app.Put("/users/me", func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}, handler.UpdateMe)
	return app
}

// TestUpdateMeUnauthorized tests that unauthenticated update requests are rejected.
// En: Ensures that update requests without authentication return a 401 error.
// Es: Asegura que las solicitudes de actualización sin autenticación devuelvan un error 401.
func TestUpdateMeUnauthorized(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	app := fiber.New()
	app.Put("/users/me", handler.UpdateMe)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeUnauthorized, errorResponse.Error.Code)
}

// TestUpdateMeNoFieldsProvided tests behavior when no update fields are provided.
// En: Verifies that an update request with no fields returns a validation error.
// Es: Verifica que una solicitud de actualización sin campos devuelva un error de validación.
func TestUpdateMeNoFieldsProvided(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	testUser := User{Name: "Original", Email: "original@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(test, handler, testUser.ID)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusBadRequest, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeValidationError, errorResponse.Error.Code)
}

// TestUpdateMeUpdateName tests successful name update.
// En: Verifies that a user can successfully update their name.
// Es: Verifica que un usuario pueda actualizar exitosamente su nombre.
func TestUpdateMeUpdateName(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	testUser := User{Name: "Old Name", Email: "updateme@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(test, handler, testUser.ID)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"New Name"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(test, "New Name", result.Data.Name)
	assert.Equal(test, testUser.Email, result.Data.Email)
	assert.Empty(test, result.Data.PasswordHash)
}

// TestUpdateMeEmailIgnored tests that email field is ignored in update requests.
// En: Ensures that email cannot be updated through the Update Me endpoint.
// Es: Asegura que el email no pueda ser actualizado a través del endpoint de actualización.
func TestUpdateMeEmailIgnored(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	testUser := User{Name: "Alice", Email: "alice.me@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(test, handler, testUser.ID)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"Alice Updated","email":"hacker@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data User `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(test, "Alice Updated", result.Data.Name)
	assert.Equal(test, "alice.me@example.com", result.Data.Email, "email must not be updated")
}

// TestUpdateMeValidationError tests validation errors during update.
// En: Verifies that validation errors are properly returned when updating with invalid data.
// Es: Verifica que los errores de validación se devuelvan correctamente al actualizar con datos inválidos.
func TestUpdateMeValidationError(test *testing.T) {
	handler := SetupUserHandlerTest(test)

	testUser := User{Name: "Bob", Email: "bob.me@example.com"}
	require.NoError(test, testUser.SetPassword("secret123"))
	require.NoError(test, database.DB.Create(&testUser).Error)

	app := setupUpdateMe(test, handler, testUser.ID)

	req := httptest.NewRequest("PUT", "/users/me", strings.NewReader(`{"name":"X","password":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnprocessableEntity, resp.StatusCode)

	errorResponse := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeValidationError, errorResponse.Error.Code)
	assert.GreaterOrEqual(test, len(errorResponse.Error.Details), 1)
}
