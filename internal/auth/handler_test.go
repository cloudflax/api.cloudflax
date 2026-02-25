package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: noopEmailSender discards all outgoing emails in tests.
// Es: noopEmailSender descarta todos los correos electrónicos salientes en pruebas.
type noopEmailSender struct{}

// En: SendTemplatedEmail does nothing in tests.
// Es: SendTemplatedEmail no hace nada en pruebas.
func (n *noopEmailSender) SendTemplatedEmail(_, _, _ string) error { return nil }

// En: handlerTestHashToken calculates the SHA-256 hash of a test token.
// Es: handlerTestHashToken calcula el hash SHA-256 de un token de prueba.
func handlerTestHashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// En: testJWTSecret is the secret key for the JWT tokens used in unit tests.
// Es: testJWTSecret es la clave secreta para los tokens JWT utilizados en pruebas unitarias.
const testJWTSecret = "test-secret-key-for-unit-tests-only"

// En: SetupAuthHandlerTest sets up the test environment and returns the auth handler and service.
// Es: SetupAuthHandlerTest configura el entorno de prueba y devuelve el manejador y servicio de autenticación.
func SetupAuthHandlerTest(test *testing.T) (*Handler, *Service) {
	test.Helper()
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	authService := NewService(authRepository, userRepository, testJWTSecret, &noopEmailSender{}, "http://test")
	authHandler := NewHandler(authService)
	return authHandler, authService
}

// En: createTestUser creates a test user.
// Es: createTestUser crea un usuario de prueba.
func createTestUser(test *testing.T, name, email, password string) *user.User {
	test.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(test, u.SetPassword(password))
	require.NoError(test, database.DB.Create(u).Error)
	return u
}

// En: createVerifiedTestUser creates a verified test user.
// Es: createVerifiedTestUser crea un usuario de prueba verificado.
func createVerifiedTestUser(test *testing.T, name, email, password string) *user.User {
	test.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(test, u.SetPassword(password))
	require.NoError(test, database.DB.Create(u).Error)
	return u
}

// En: DecodeErrorResponse decodes the response body into a runtimeError.ErrorResponse.
// Es: DecodeErrorResponse decodifica el cuerpo de la respuesta en un runtimeError.ErrorResponse.
func DecodeErrorResponse(test *testing.T, body io.Reader) runtimeError.ErrorResponse {
	test.Helper()
	var result runtimeError.ErrorResponse
	require.NoError(test, json.NewDecoder(body).Decode(&result))
	return result
}

// --- Login ---

// En: TestLoginSuccess tests the successful login.
// Es: TestLoginSuccess prueba el inicio de sesión exitoso.
func TestLoginSuccess(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)
	createVerifiedTestUser(test, "Alice", "alice@example.com", "password123")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"alice@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data TokenPair `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(test, result.Data.AccessToken)
	assert.NotEmpty(test, result.Data.RefreshToken)
	assert.False(test, result.Data.ExpiresAt.IsZero())
}

// En: TestLoginInvalidPassword tests the login with an invalid password.
// Es: TestLoginInvalidPassword prueba el inicio de sesión con contraseña incorrecta.
func TestLoginInvalidPassword(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)
	createTestUser(test, "Bob", "bob@example.com", "correctpassword")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"bob@example.com","password":"wrongpassword"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeInvalidCredentials, errResp.Error.Code)
}

// En: TestLoginUserNotFound tests the login with a non-existent email.
// Es: TestLoginUserNotFound prueba el inicio de sesión con correo electrónico no encontrado.
func TestLoginUserNotFound(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"nobody@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeInvalidCredentials, errResp.Error.Code)
}

// En: TestLoginEmailNotVerified tests the login with an unverified email.
// Es: TestLoginEmailNotVerified prueba el inicio de sesión con correo electrónico no verificado.
func TestLoginEmailNotVerified(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)
	createTestUser(test, "Unverified", "unverified@example.com", "password123")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"unverified@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusForbidden, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeEmailVerificationRequired, errResp.Error.Code)
	assert.Contains(test, errResp.Error.Message, "verified")
}

// En: TestLoginValidationError tests the login with a validation error.
// Es: TestLoginValidationError prueba el inicio de sesión con un error de validación.
func TestLoginValidationError(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"not-an-email","password":"short"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeValidationError, errResp.Error.Code)
	assert.NotEmpty(test, errResp.Error.Details)
}

// En: TestLoginCaseInsensitiveEmail tests the login with a case-insensitive email.
// Es: TestLoginCaseInsensitiveEmail prueba el inicio de sesión con correo electrónico insensible a mayúsculas y minúsculas.
func TestLoginCaseInsensitiveEmail(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)
	createVerifiedTestUser(test, "Carol", "carol@example.com", "password123")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"CAROL@EXAMPLE.COM","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)
}

// --- Refresh ---

// En: TestRefreshSuccess tests the successful refresh.
// Es: TestRefreshSuccess prueba el refresco de token exitoso.
func TestRefreshSuccess(test *testing.T) {
	handler, service := SetupAuthHandlerTest(test)
	createVerifiedTestUser(test, "Dave", "dave@example.com", "password123")

	pair, err := service.Login("dave@example.com", "password123")
	require.NoError(test, err)

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": pair.RefreshToken})
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data TokenPair `json:"data"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(test, result.Data.AccessToken)
	assert.NotEmpty(test, result.Data.RefreshToken)
	assert.NotEqual(test, pair.RefreshToken, result.Data.RefreshToken, "refresh token must be rotated")
}

// En: TestRefreshInvalidToken tests the refresh with an invalid token.
// Es: TestRefreshInvalidToken prueba el refresco de token con token inválido.
func TestRefreshInvalidToken(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": "invalid-token-value"})
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeTokenInvalid, errResp.Error.Code)
}

// En: TestRefreshTokenRotationOldTokenInvalidAfterRefresh tests the refresh token rotation with an invalid token after refresh.
// Es: TestRefreshTokenRotationOldTokenInvalidAfterRefresh prueba el refresco de token con token inválido después de refrescar.
func TestRefreshTokenRotationOldTokenInvalidAfterRefresh(test *testing.T) {
	handler, service := SetupAuthHandlerTest(test)
	createVerifiedTestUser(test, "Eve", "eve@example.com", "password123")

	pair, err := service.Login("eve@example.com", "password123")
	require.NoError(test, err)

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": pair.RefreshToken})

	// First refresh — should succeed.
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	resp.Body.Close()
	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	// Second refresh with the same (now revoked) token — must be rejected.
	req2 := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp2.Body.Close()
	assert.Equal(test, fiber.StatusUnauthorized, resp2.StatusCode)
}

// En: TestRefreshEmailNotVerified tests the refresh with an unverified email.
// Es: TestRefreshEmailNotVerified prueba el refresco de token con correo electrónico no verificado.
func TestRefreshEmailNotVerified(test *testing.T) {
	handler, service := SetupAuthHandlerTest(test)
	u := createTestUser(test, "Unverified", "unverified@example.com", "password123")

	rawToken := "refresh-token-unverified-user"
	require.NoError(test, service.repository.Create(&RefreshToken{
		UserID:    u.ID,
		TokenHash: handlerTestHashToken(rawToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}))

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": rawToken})
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusForbidden, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeEmailVerificationRequired, errResp.Error.Code)
}

// --- Logout ---

// En: TestLogoutSuccess tests the successful logout.
// Es: TestLogoutSuccess prueba el cierre de sesión exitoso.
func TestLogoutSuccess(test *testing.T) {
	handler, service := SetupAuthHandlerTest(test)
	createVerifiedTestUser(test, "Frank", "frank@example.com", "password123")

	pair, err := service.Login("frank@example.com", "password123")
	require.NoError(test, err)

	app := fiber.New()
	// Simulate the auth middleware by manually setting userID in locals.
	app.Post("/auth/logout", func(c fiber.Ctx) error {
		userID, _, parseErr := service.ValidateAccessToken(pair.AccessToken)
		if parseErr != nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		c.Locals("userID", userID)
		return c.Next()
	}, handler.Logout)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusNoContent, resp.StatusCode)
}

// En: TestLogoutWithoutAuth tests the logout without authentication.
// Es: TestLogoutWithoutAuth prueba el cierre de sesión sin autenticación.
func TestLogoutWithoutAuth(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/logout", handler.Logout)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnauthorized, resp.StatusCode)
}

// --- Register ---

// En: TestRegisterSuccess tests the successful registration.
// Es: TestRegisterSuccess prueba el registro exitoso.
func TestRegisterSuccess(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/register", handler.Register)

	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data user.User `json:"data"`
		Meta struct {
			EmailVerificationRequired bool `json:"email_verification_required"`
		} `json:"meta"`
	}
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(test, result.Data.ID)
	assert.Equal(test, "Alice", result.Data.Name)
	assert.Equal(test, "alice@example.com", result.Data.Email)
	assert.Empty(test, result.Data.PasswordHash)
	assert.Nil(test, result.Data.EmailVerifiedAt)
	assert.True(test, result.Meta.EmailVerificationRequired)

	var provider UserAuthProvider
	dbErr := database.DB.Where("user_id = ? AND provider = ?", result.Data.ID, ProviderCredentials).First(&provider).Error
	require.NoError(test, dbErr, "UserAuthProvider must be created")
	assert.Equal(test, "alice@example.com", provider.ProviderSubjectID)
}

// En: TestRegisterDuplicateEmail tests the registration with a duplicate email.
// Es: TestRegisterDuplicateEmail prueba el registro con correo electrónico duplicado.
func TestRegisterDuplicateEmail(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/register", handler.Register)

	body := strings.NewReader(`{"name":"Alice","email":"dup@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	resp.Body.Close()
	require.Equal(test, fiber.StatusCreated, resp.StatusCode)

	body2 := strings.NewReader(`{"name":"Bob","email":"dup@example.com","password":"password456"}`)
	req2 := httptest.NewRequest("POST", "/auth/register", body2)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp2.Body.Close()

	assert.Equal(test, fiber.StatusConflict, resp2.StatusCode)
	errResp := DecodeErrorResponse(test, resp2.Body)
	assert.Equal(test, runtimeError.CodeEmailAlreadyExists, errResp.Error.Code)
}

// En: TestRegisterValidationError tests the registration with a validation error.
// Es: TestRegisterValidationError prueba el registro con un error de validación.
func TestRegisterValidationError(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/register", handler.Register)

	body := strings.NewReader(`{"name":"A","email":"not-email","password":"short"}`)
	req := httptest.NewRequest("POST", "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeValidationError, errResp.Error.Code)
}

// --- VerifyEmail ---

// En: TestVerifyEmailSuccess tests the successful email verification.
// Es: TestVerifyEmailSuccess prueba la verificación de correo electrónico exitosa.
func TestVerifyEmailSuccess(test *testing.T) {
	handler, authService := SetupAuthHandlerTest(test)

	_, token, err := authService.Register("Carol", "carol@example.com", "password123")
	require.NoError(test, err)
	require.NotEmpty(test, token)

	app := fiber.New()
	app.Get("/auth/verify-email", handler.VerifyEmail)

	req := httptest.NewRequest("GET", "/auth/verify-email?token="+token, nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var u user.User
	require.NoError(test, database.DB.Where("email = ?", "carol@example.com").First(&u).Error)
	assert.NotNil(test, u.EmailVerifiedAt, "email_verified_at must be set")
	assert.Nil(test, u.EmailVerificationToken, "verification token must be cleared")
}

// En: TestVerifyEmailInvalidToken tests the email verification with an invalid token.
// Es: TestVerifyEmailInvalidToken prueba la verificación de correo electrónico con token inválido.
func TestVerifyEmailInvalidToken(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Get("/auth/verify-email", handler.VerifyEmail)

	req := httptest.NewRequest("GET", "/auth/verify-email?token=00000000-0000-0000-0000-000000000000", nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeInvalidVerificationToken, errResp.Error.Code)
}

// --- ResendVerification ---

// En: TestResendVerificationSuccess tests the successful email verification resend.
// Es: TestResendVerificationSuccess prueba el reenvío de verificación de correo electrónico exitoso.
func TestResendVerificationSuccess(test *testing.T) {
	handler, authService := SetupAuthHandlerTest(test)

	_, _, err := authService.Register("Dan", "dan@example.com", "password123")
	require.NoError(test, err)

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"dan@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode)

	var u user.User
	require.NoError(test, database.DB.Where("email = ?", "dan@example.com").First(&u).Error)
	assert.NotNil(test, u.EmailVerificationToken, "a new token must be set")
}

// En: TestResendVerificationAlreadyVerified tests the email verification resend with an already verified email.
// Es: TestResendVerificationAlreadyVerified prueba el reenvío de verificación de correo electrónico con correo electrónico ya verificado.
func TestResendVerificationAlreadyVerified(test *testing.T) {
	handler, authService := SetupAuthHandlerTest(test)

	_, token, err := authService.Register("Eve", "eve@example.com", "password123")
	require.NoError(test, err)
	require.NoError(test, authService.VerifyEmail(token))

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"eve@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusConflict, resp.StatusCode)
	errResp := DecodeErrorResponse(test, resp.Body)
	assert.Equal(test, runtimeError.CodeEmailAlreadyVerified, errResp.Error.Code)
}

// En: TestResendVerificationUnknownEmail tests the email verification resend with an unknown email.
// Es: TestResendVerificationUnknownEmail prueba el reenvío de verificación de correo electrónico con correo electrónico desconocido.
func TestResendVerificationUnknownEmail(test *testing.T) {
	handler, _ := SetupAuthHandlerTest(test)

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"ghost@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusOK, resp.StatusCode, "unknown email must return 200 to prevent enumeration")
}
