package invoice

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	apierrors "github.com/cloudflax/api.cloudflax/internal/shared/errors"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlerTest(t *testing.T) (*Handler, *account.Account) {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &account.Account{}, &account.AccountMember{}, &Invoice{}))

	acc := &account.Account{Name: "Test Co", Slug: "test-co"}
	require.NoError(t, database.DB.Create(acc).Error)

	repository := NewRepository(database.DB)
	service := NewService(repository)
	return NewHandler(service), acc
}

func decodeErrorResponse(t *testing.T, body io.Reader) apierrors.ErrorResponse {
	t.Helper()
	var result apierrors.ErrorResponse
	require.NoError(t, json.NewDecoder(body).Decode(&result))
	return result
}

func injectContext(userID, accountID string) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		c.Locals("accountID", accountID)
		return c.Next()
	}
}

func TestHandler_ListInvoice_Success(t *testing.T) {
	handler, acc := setupHandlerTest(t)

	repository := NewRepository(database.DB)
	inv := &Invoice{AccountID: acc.ID, Number: "INV-001", Status: StatusDraft, TotalCents: 1000, Currency: "USD"}
	require.NoError(t, repository.CreateInvoice(inv))

	app := fiber.New()
	app.Get("/invoices", injectContext("user-1", acc.ID), handler.ListInvoice)

	req := httptest.NewRequest("GET", "/invoices", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result struct {
		Data []Invoice `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Len(t, result.Data, 1)
}

func TestHandler_ListInvoice_Unauthorized(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	app := fiber.New()
	app.Get("/invoices", handler.ListInvoice)

	req := httptest.NewRequest("GET", "/invoices", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestHandler_GetInvoice_Success(t *testing.T) {
	handler, acc := setupHandlerTest(t)

	repository := NewRepository(database.DB)
	inv := &Invoice{AccountID: acc.ID, Number: "INV-001", Status: StatusDraft, TotalCents: 1000, Currency: "USD"}
	require.NoError(t, repository.CreateInvoice(inv))

	app := fiber.New()
	app.Get("/invoices/:id", injectContext("user-1", acc.ID), handler.GetInvoice)

	req := httptest.NewRequest("GET", "/invoices/"+inv.ID, nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_GetInvoice_NotFound(t *testing.T) {
	handler, acc := setupHandlerTest(t)

	app := fiber.New()
	app.Get("/invoices/:id", injectContext("user-1", acc.ID), handler.GetInvoice)

	req := httptest.NewRequest("GET", "/invoices/00000000-0000-0000-0000-000000000000", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeInvoiceNotFound, errResp.Error.Code)
}

func TestHandler_CreateInvoice_Success(t *testing.T) {
	handler, acc := setupHandlerTest(t)

	app := fiber.New()
	app.Post("/invoices", injectContext("user-1", acc.ID), handler.CreateInvoice)

	body := `{"number":"INV-001","currency":"USD","total_cents":9900}`
	req := httptest.NewRequest("POST", "/invoices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var result struct {
		Data Invoice `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.Data.ID)
	assert.Equal(t, acc.ID, result.Data.AccountID)
	assert.Equal(t, "INV-001", result.Data.Number)
	assert.Equal(t, int64(9900), result.Data.TotalCents)
}

func TestHandler_CreateInvoice_ValidationError(t *testing.T) {
	handler, acc := setupHandlerTest(t)

	app := fiber.New()
	app.Post("/invoices", injectContext("user-1", acc.ID), handler.CreateInvoice)

	body := `{"number":"","currency":"USD","total_cents":0}`
	req := httptest.NewRequest("POST", "/invoices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
	errResp := decodeErrorResponse(t, resp.Body)
	assert.Equal(t, apierrors.CodeValidationError, errResp.Error.Code)
}

func TestHandler_CreateInvoice_Unauthorized(t *testing.T) {
	handler, _ := setupHandlerTest(t)

	app := fiber.New()
	app.Post("/invoices", handler.CreateInvoice)

	body := `{"number":"INV-001","currency":"USD","total_cents":1000}`
	req := httptest.NewRequest("POST", "/invoices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
