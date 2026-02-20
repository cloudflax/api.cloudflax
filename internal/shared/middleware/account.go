package middleware

import (
	"errors"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/gofiber/fiber/v3"
)

// AccountRepository is the interface the account middleware needs to resolve
// account identity and validate membership.
type AccountRepository interface {
	GetByID(id string) (*account.Account, error)
	GetBySlug(slug string) (*account.Account, error)
	GetMember(accountID, userID string) (*account.AccountMember, error)
}

// RequireAccountMember returns a Fiber middleware that resolves the target account
// from the request and verifies that the authenticated user is a member.
//
// The account is identified by (in order of precedence):
//  1. X-Account-ID header
//  2. account_id query parameter
//  3. X-Account-Slug header
//  4. account_slug query parameter
//
// On success it sets "accountID" in Fiber locals and calls Next.
// It requires RequireAuth to run first (userID must already be in locals).
func RequireAccountMember(repo AccountRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return runtimeerror.Respond(c, fiber.StatusUnauthorized, runtimeerror.CodeUnauthorized, "Unauthorized")
		}

		acc, err := resolveAccount(c, repo)
		if err != nil {
			if errors.Is(err, account.ErrNotFound) {
				return runtimeerror.Respond(c, fiber.StatusNotFound, runtimeerror.CodeAccountNotFound, "Account not found")
			}
			return runtimeerror.Respond(c, fiber.StatusBadRequest, runtimeerror.CodeInvalidRequestBody, "Account identifier required (X-Account-ID, X-Account-Slug, account_id or account_slug)")
		}

		if _, err := repo.GetMember(acc.ID, userID); err != nil {
			if errors.Is(err, account.ErrMemberNotFound) {
				return runtimeerror.Respond(c, fiber.StatusForbidden, runtimeerror.CodeForbidden, "Access denied: not a member of this account")
			}
			return runtimeerror.Respond(c, fiber.StatusInternalServerError, runtimeerror.CodeInternalServerError, "Failed to verify account membership")
		}

		c.Locals("accountID", acc.ID)
		return c.Next()
	}
}

// resolveAccount finds the account from the request headers or query params.
// Returns errNoAccountIdentifier if neither ID nor slug is provided.
func resolveAccount(c fiber.Ctx, repo AccountRepository) (*account.Account, error) {
	if id := firstNonEmpty(c.Get("X-Account-ID"), c.Query("account_id")); id != "" {
		return repo.GetByID(id)
	}

	if slug := firstNonEmpty(c.Get("X-Account-Slug"), c.Query("account_slug")); slug != "" {
		return repo.GetBySlug(slug)
	}

	return nil, errNoAccountIdentifier
}

// errNoAccountIdentifier is returned when no account identifier is present in the request.
var errNoAccountIdentifier = errors.New("no account identifier in request")

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
