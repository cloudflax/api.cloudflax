package auth

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key-for-unit-tests-only"

func setupAuthHandlerTest(t *testing.T) (*Handler, *Service) {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	authService := NewService(authRepository, userRepository, testJWTSecret)
	authHandler := NewHandler(authService)
	return authHandler, authService
}

func createTestUser(t *testing.T, name, email, password string) *user.User {
	t.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(t, u.SetPassword(password))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func decodeErrorResponse(t *testing.T, body io.Reader) runtimeerror.ErrorResponse {
	t.Helper()
	var result runtimeerror.ErrorResponse
	require.NoError(t, json.NewDecoder(body).Decode(&result))
	return result
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)
	createTestUser(t, "Alice", "alice@example.com", "password123")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"alice@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data TokenPair `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Data.AccessToken)
	assert.NotEmpty(t, result.Data.RefreshToken)
	assert.False(t, result.Data.ExpiresAt.IsZero())
}

func TestLogin_InvalidPassword(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)
	createTestUser(t, "Bob", "bob@example.com", "correctpassword")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"bob@example.com","password":"wrongpassword"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeInvalidCredentials, errResp.Error.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"nobody@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeInvalidCredentials, errResp.Error.Code)
}

func TestLogin_ValidationError(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"not-an-email","password":"short"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
	assert.NotEmpty(t, errResp.Error.Details)
}

func TestLogin_CaseInsensitiveEmail(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)
	createTestUser(t, "Carol", "carol@example.com", "password123")

	app := fiber.New()
	app.Post("/auth/login", handler.Login)

	body := strings.NewReader(`{"email":"CAROL@EXAMPLE.COM","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	handler, service := setupAuthHandlerTest(t)
	createTestUser(t, "Dave", "dave@example.com", "password123")

	pair, err := service.Login("dave@example.com", "password123")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": pair.RefreshToken})
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data TokenPair `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Data.AccessToken)
	assert.NotEmpty(t, result.Data.RefreshToken)
	assert.NotEqual(t, pair.RefreshToken, result.Data.RefreshToken, "refresh token must be rotated")
}

func TestRefresh_InvalidToken(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": "invalid-token-value"})
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeTokenInvalid, errResp.Error.Code)
}

func TestRefresh_TokenRotation_OldTokenInvalidAfterRefresh(t *testing.T) {
	handler, service := setupAuthHandlerTest(t)
	createTestUser(t, "Eve", "eve@example.com", "password123")

	pair, err := service.Login("eve@example.com", "password123")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/refresh", handler.Refresh)

	bodyStr, _ := json.Marshal(map[string]string{"refresh_token": pair.RefreshToken})

	// First refresh — should succeed.
	req := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Second refresh with the same (now revoked) token — must be rejected.
	req2 := httptest.NewRequest("POST", "/auth/refresh", strings.NewReader(string(bodyStr)))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp2.StatusCode)
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	handler, service := setupAuthHandlerTest(t)
	createTestUser(t, "Frank", "frank@example.com", "password123")

	pair, err := service.Login("frank@example.com", "password123")
	require.NoError(t, err)

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
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestLogout_WithoutAuth(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/logout", handler.Logout)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/register", handler.Register)

	body := strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data user.User `json:"data"`
		Meta struct {
			EmailVerificationRequired bool `json:"email_verification_required"`
		} `json:"meta"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Data.ID)
	assert.Equal(t, "Alice", result.Data.Name)
	assert.Equal(t, "alice@example.com", result.Data.Email)
	assert.Empty(t, result.Data.PasswordHash)
	assert.Nil(t, result.Data.EmailVerifiedAt)
	assert.True(t, result.Meta.EmailVerificationRequired)

	var provider UserAuthProvider
	dbErr := database.DB.Where("user_id = ? AND provider = ?", result.Data.ID, ProviderCredentials).First(&provider).Error
	require.NoError(t, dbErr, "UserAuthProvider must be created")
	assert.Equal(t, "alice@example.com", provider.ProviderSubjectID)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/register", handler.Register)

	body := strings.NewReader(`{"name":"Alice","email":"dup@example.com","password":"password123"}`)
	req := httptest.NewRequest("POST", "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body2 := strings.NewReader(`{"name":"Bob","email":"dup@example.com","password":"password456"}`)
	req2 := httptest.NewRequest("POST", "/auth/register", body2)
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, fiber.StatusConflict, resp2.StatusCode)
	errResp := decodeErrorResponse(t, resp2.Body)
	assert.Equal(t, runtimeerror.CodeEmailAlreadyExists, errResp.Error.Code)
}

func TestRegister_ValidationError(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/register", handler.Register)

	body := strings.NewReader(`{"name":"A","email":"not-email","password":"short"}`)
	req := httptest.NewRequest("POST", "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
}

// --- VerifyEmail ---

func TestVerifyEmail_Success(t *testing.T) {
	handler, authService := setupAuthHandlerTest(t)

	_, token, err := authService.Register("Carol", "carol@example.com", "password123")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	app := fiber.New()
	app.Post("/auth/verify-email", handler.VerifyEmail)

	body := strings.NewReader(`{"token":"` + token + `"}`)
	req := httptest.NewRequest("POST", "/auth/verify-email", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var u user.User
	require.NoError(t, database.DB.Where("email = ?", "carol@example.com").First(&u).Error)
	assert.NotNil(t, u.EmailVerifiedAt, "email_verified_at must be set")
	assert.Nil(t, u.EmailVerificationToken, "verification token must be cleared")
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/verify-email", handler.VerifyEmail)

	body := strings.NewReader(`{"token":"00000000-0000-0000-0000-000000000000"}`)
	req := httptest.NewRequest("POST", "/auth/verify-email", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeInvalidVerificationToken, errResp.Error.Code)
}

// --- ResendVerification ---

func TestResendVerification_Success(t *testing.T) {
	handler, authService := setupAuthHandlerTest(t)

	_, _, err := authService.Register("Dan", "dan@example.com", "password123")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"dan@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var u user.User
	require.NoError(t, database.DB.Where("email = ?", "dan@example.com").First(&u).Error)
	assert.NotNil(t, u.EmailVerificationToken, "a new token must be set")
}

func TestResendVerification_AlreadyVerified(t *testing.T) {
	handler, authService := setupAuthHandlerTest(t)

	_, token, err := authService.Register("Eve", "eve@example.com", "password123")
	require.NoError(t, err)
	require.NoError(t, authService.VerifyEmail(token))

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"eve@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeEmailAlreadyVerified, errResp.Error.Code)
}

func TestResendVerification_UnknownEmail(t *testing.T) {
	handler, _ := setupAuthHandlerTest(t)

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"ghost@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "unknown email must return 200 to prevent enumeration")
}
