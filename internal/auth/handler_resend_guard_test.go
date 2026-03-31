package auth

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResendGuard struct {
	err error
}

func (g testResendGuard) CheckAndConsume(_ context.Context, _, _ string) error {
	return g.err
}

// En: TestResendVerificationRateLimited tests resend endpoint throttle response.
// Es: TestResendVerificationRateLimited prueba respuesta throttle del endpoint resend.
func TestResendVerificationRateLimited(test *testing.T) {
	handler, authService := SetupAuthHandlerTest(test)
	_, _, err := authService.Register("Rate", "rate@example.com", "password123")
	require.NoError(test, err)

	handler.WithResendVerificationGuard(testResendGuard{
		err: &ResendVerificationRateLimitError{RetryAfter: 2 * time.Hour},
	})

	app := fiber.New()
	app.Post("/auth/resend-verification", handler.ResendVerification)

	body := strings.NewReader(`{"email":"rate@example.com"}`)
	req := httptest.NewRequest("POST", "/auth/resend-verification", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0})
	require.NoError(test, err)
	defer resp.Body.Close()

	assert.Equal(test, fiber.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(test, "7200", resp.Header.Get("Retry-After"))

	var result runtimeError.ErrorResponse
	require.NoError(test, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(test, runtimeError.CodeRateLimited, result.Error.Code)
}
