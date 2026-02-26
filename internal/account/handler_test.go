package account

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlerTest(t *testing.T) (*Handler, *user.Repository) {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &Account{}, &AccountMember{}))

	userRepository := user.NewRepository(database.DB)
	accountRepository := NewRepository(database.DB)
	service := NewService(accountRepository, userRepository)
	return NewHandler(service), userRepository
}

func decodeErrorResponse(t *testing.T, body io.Reader) runtimeerror.ErrorResponse {
	t.Helper()
	var result runtimeerror.ErrorResponse
	require.NoError(t, json.NewDecoder(body).Decode(&result))
	return result
}

func seedVerifiedUserForHandler(t *testing.T, name, email string) *user.User {
	t.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(t, u.SetPassword("password123"))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func seedUnverifiedUserForHandler(t *testing.T, name, email string) *user.User {
	t.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(t, u.SetPassword("password123"))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func TestCreateAccount_Success(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Alice", "alice@example.com")

	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Acme"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data Account `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Data.ID)
	assert.Equal(t, "Acme", result.Data.Name)
	assert.Equal(t, "acme", result.Data.Slug)
}

func TestCreateAccount_WithCustomSlug(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Bob", "bob@example.com")

	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Bob Corp","slug":"bob-corp"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data Account `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "bob-corp", result.Data.Slug)
}

func TestCreateAccount_Unauthorized(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	app := fiber.New()
	app.Post("/accounts", handler.CreateAccount)

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Acme"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeUnauthorized, errResp.Error.Code)
}

func TestCreateAccount_ValidationError(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Carol", "carol@example.com")

	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
	assert.NotEmpty(t, errResp.Error.Details)
}

func TestCreateAccount_InvalidSlugFormat(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Dave", "dave@example.com")

	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Dave Corp","slug":"Invalid Slug!"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeValidationError, errResp.Error.Code)
}

func TestCreateAccount_EmailNotVerified(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedUnverifiedUserForHandler(t, "Eve", "eve@example.com")

	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Eve Org"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeEmailVerificationRequired, errResp.Error.Code)
}

func TestCreateAccount_SlugTaken(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Frank", "frank@example.com")

	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	firstReq := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"First","slug":"taken-slug"}`))
	firstReq.Header.Set("Content-Type", "application/json")
	firstResp, err := app.Test(firstReq, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	firstResp.Body.Close()
	require.Equal(t, fiber.StatusCreated, firstResp.StatusCode)

	secondReq := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Second","slug":"taken-slug"}`))
	secondReq.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(secondReq, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, runtimeerror.CodeAccountSlugTaken, errResp.Error.Code)
}

func TestSetActiveAccount_Success(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Grace", "grace@example.com")

	// First create an account for the owner
	app := fiber.New()
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	createReq := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Grace Org"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := app.Test(createReq, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, fiber.StatusCreated, createResp.StatusCode)

	var created struct {
		Data Account `json:"data"`
	}
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))

	// Now mount the active account endpoint
	app.Post("/accounts/active", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.SetActiveAccount)

	activeReqBody := `{"account_id":"` + created.Data.ID + `"}`
	activeReq := httptest.NewRequest("POST", "/accounts/active", strings.NewReader(activeReqBody))
	activeReq.Header.Set("Content-Type", "application/json")

	activeResp, err := app.Test(activeReq, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer activeResp.Body.Close()

	assert.Equal(t, fiber.StatusOK, activeResp.StatusCode)

	var activeResult struct {
		Data struct {
			ActiveAccountID string `json:"active_account_id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(activeResp.Body).Decode(&activeResult))
	assert.Equal(t, created.Data.ID, activeResult.Data.ActiveAccountID)
}

func TestSetActiveAccount_NotMember(t *testing.T) {
	handler, _ := setupHandlerTest(t)
	owner := seedVerifiedUserForHandler(t, "Kelly", "kelly@example.com")
	other := seedVerifiedUserForHandler(t, "Liam", "liam@example.com")

	app := fiber.New()

	// Create account for owner
	app.Post("/accounts", func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID)
		return c.Next()
	}, handler.CreateAccount)

	createReq := httptest.NewRequest("POST", "/accounts", strings.NewReader(`{"name":"Owner Org"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := app.Test(createReq, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, fiber.StatusCreated, createResp.StatusCode)

	var created struct {
		Data Account `json:"data"`
	}
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))

	// Mount active endpoint, simulating authenticated as "other"
	app.Post("/accounts/active", func(c fiber.Ctx) error {
		c.Locals("userID", other.ID)
		return c.Next()
	}, handler.SetActiveAccount)

	activeReqBody := `{"account_id":"` + created.Data.ID + `"}`
	activeReq := httptest.NewRequest("POST", "/accounts/active", strings.NewReader(activeReqBody))
	activeReq.Header.Set("Content-Type", "application/json")

	activeResp, err := app.Test(activeReq, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer activeResp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, activeResp.StatusCode)
	errResp := decodeErrorResponse(t, activeResp.Body)
	assert.Equal(t, runtimeerror.CodeForbidden, errResp.Error.Code)
}
