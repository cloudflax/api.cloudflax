package errors

import "github.com/gofiber/fiber/v3"

// ErrorCode is a stable, machine-readable error identifier.
type ErrorCode string

// Generic error codes.
const (
	CodeInvalidRequestBody  ErrorCode = "INVALID_REQUEST_BODY"
	CodeValidationError     ErrorCode = "VALIDATION_ERROR"
	CodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
)

// User error codes.
const (
	CodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	CodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"
)

// Auth error codes.
const (
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	CodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
)

// ErrorDetail describes a single field-level validation failure.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIError is the canonical error payload returned by the API.
type APIError struct {
	Code    ErrorCode     `json:"code"`
	Message string        `json:"message"`
	Status  int           `json:"status"`
	TraceID string        `json:"trace_id,omitempty"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorResponse wraps APIError as the top-level JSON response.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// Respond writes a single-error JSON response and returns the Fiber error.
func Respond(c fiber.Ctx, status int, code ErrorCode, message string) error {
	traceID := c.Get("X-Request-ID")
	return c.Status(status).JSON(ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Status:  status,
			TraceID: traceID,
		},
	})
}

// RespondWithDetails writes a multi-detail JSON error response (e.g. validation failures).
func RespondWithDetails(c fiber.Ctx, status int, code ErrorCode, message string, details []ErrorDetail) error {
	traceID := c.Get("X-Request-ID")
	return c.Status(status).JSON(ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Status:  status,
			TraceID: traceID,
			Details: details,
		},
	})
}
