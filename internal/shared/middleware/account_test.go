package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAccountMiddlewareTest(t *testing.T) (*account.Repository, *user.Repository) {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &account.Account{}, &account.AccountMember{}))
	return account.NewRepository(database.DB), user.NewRepository(database.DB)
}

func seedVerifiedUser(t *testing.T, repo *user.Repository, name, email string) *user.User {
	t.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(t, u.SetPassword("pass"))
	require.NoError(t, repo.Create(u))
	return u
}

func seedAccountWithOwner(t *testing.T, repo *account.Repository, ownerID, name, slug string) *account.Account {
	t.Helper()
	acc := &account.Account{Name: name, Slug: slug}
	require.NoError(t, repo.CreateAccount(acc))
	member := &account.AccountMember{AccountID: acc.ID, UserID: ownerID, Role: account.RoleOwner}
	require.NoError(t, repo.CreateMember(member))
	return acc
}

func injectUserID(userID string) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	}
}

func newAppWithMiddleware(repo AccountRepository, userID string) *fiber.App {
	app := fiber.New()
	app.Get("/test",
		injectUserID(userID),
		RequireAccountMember(repo),
		func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"accountID": c.Locals("accountID")})
		},
	)
	return app
}

func decodeAccountErrorResponse(t *testing.T, resp *httptest.ResponseRecorder) runtimeerror.ErrorResponse {
	t.Helper()
	var result runtimeerror.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func TestRequireAccountMember_ByIDHeader_Success(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Alice", "alice@example.com")
	acc := seedAccountWithOwner(t, accountRepo, owner.ID, "Acme", "acme")

	app := newAppWithMiddleware(accountRepo, owner.ID)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-ID", acc.ID)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAccountMember_BySlugHeader_Success(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Bob", "bob@example.com")
	seedAccountWithOwner(t, accountRepo, owner.ID, "Bob Corp", "bob-corp")

	app := newAppWithMiddleware(accountRepo, owner.ID)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-Slug", "bob-corp")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAccountMember_ByIDQueryParam_Success(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Carol", "carol@example.com")
	acc := seedAccountWithOwner(t, accountRepo, owner.ID, "Carol Co", "carol-co")

	app := newAppWithMiddleware(accountRepo, owner.ID)
	req := httptest.NewRequest("GET", "/test?account_id="+acc.ID, nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAccountMember_BySlugQueryParam_Success(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Dave", "dave@example.com")
	seedAccountWithOwner(t, accountRepo, owner.ID, "Dave Inc", "dave-inc")

	app := newAppWithMiddleware(accountRepo, owner.ID)
	req := httptest.NewRequest("GET", "/test?account_slug=dave-inc", nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestRequireAccountMember_InjectsAccountIDInLocals(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Eve", "eve@example.com")
	acc := seedAccountWithOwner(t, accountRepo, owner.ID, "Eve LLC", "eve-llc")

	var capturedAccountID string
	app := fiber.New()
	app.Get("/test",
		injectUserID(owner.ID),
		RequireAccountMember(accountRepo),
		func(c fiber.Ctx) error {
			capturedAccountID, _ = c.Locals("accountID").(string)
			return c.SendStatus(fiber.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-ID", acc.ID)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, acc.ID, capturedAccountID)
}

func TestRequireAccountMember_NoIdentifier_BadRequest(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Frank", "frank@example.com")

	app := newAppWithMiddleware(accountRepo, owner.ID)
	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestRequireAccountMember_AccountNotFound_404(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Grace", "grace@example.com")

	app := newAppWithMiddleware(accountRepo, owner.ID)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-ID", "00000000-0000-0000-0000-000000000000")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var result runtimeerror.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, runtimeerror.CodeAccountNotFound, result.Error.Code)
}

func TestRequireAccountMember_NotMember_Forbidden(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Heidi", "heidi@example.com")
	outsider := seedVerifiedUser(t, userRepo, "Ivan", "ivan@example.com")
	acc := seedAccountWithOwner(t, accountRepo, owner.ID, "Heidi Co", "heidi-co")

	app := newAppWithMiddleware(accountRepo, outsider.ID)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-ID", acc.ID)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var result runtimeerror.ErrorResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, runtimeerror.CodeForbidden, result.Error.Code)
}

func TestRequireAccountMember_NoUserID_Unauthorized(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Judy", "judy@example.com")
	acc := seedAccountWithOwner(t, accountRepo, owner.ID, "Judy Inc", "judy-inc")

	app := fiber.New()
	app.Get("/test", RequireAccountMember(accountRepo), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-ID", acc.ID)

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAccountMember_IDTakesPrecedenceOverSlug(t *testing.T) {
	accountRepo, userRepo := setupAccountMiddlewareTest(t)
	owner := seedVerifiedUser(t, userRepo, "Karl", "karl@example.com")
	accByID := seedAccountWithOwner(t, accountRepo, owner.ID, "Karl A", "karl-a")
	seedAccountWithOwner(t, accountRepo, owner.ID, "Karl B", "karl-b")

	var capturedAccountID string
	app := fiber.New()
	app.Get("/test",
		injectUserID(owner.ID),
		RequireAccountMember(accountRepo),
		func(c fiber.Ctx) error {
			capturedAccountID, _ = c.Locals("accountID").(string)
			return c.SendStatus(fiber.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Account-ID", accByID.ID)
	req.Header.Set("X-Account-Slug", "karl-b")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, accByID.ID, capturedAccountID)
}
