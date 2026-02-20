package requestctx

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// ErrMissingUserID is returned when the userID is not present in Fiber locals.
var ErrMissingUserID = errors.New("userID not found in request context")

// ErrMissingAccountID is returned when the accountID is not present in Fiber locals.
var ErrMissingAccountID = errors.New("accountID not found in request context")

// RequestContext holds the authenticated identity extracted from Fiber locals.
// It is populated by the RequireAuth and RequireAccountMember middlewares.
type RequestContext struct {
	UserID    string
	Email     string
	AccountID string
}

// FromFiber extracts a full RequestContext from Fiber locals.
// Requires both "userID" (set by RequireAuth) and "accountID" (set by RequireAccountMember).
func FromFiber(c fiber.Ctx) (*RequestContext, error) {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return nil, ErrMissingUserID
	}

	accountID, ok := c.Locals("accountID").(string)
	if !ok || accountID == "" {
		return nil, ErrMissingAccountID
	}

	email, _ := c.Locals("email").(string)

	return &RequestContext{
		UserID:    userID,
		Email:     email,
		AccountID: accountID,
	}, nil
}

// UserOnly extracts a RequestContext without requiring AccountID.
// Use on routes that require authentication but do not operate within an account scope.
func UserOnly(c fiber.Ctx) (*RequestContext, error) {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return nil, ErrMissingUserID
	}

	email, _ := c.Locals("email").(string)

	return &RequestContext{
		UserID: userID,
		Email:  email,
	}, nil
}
