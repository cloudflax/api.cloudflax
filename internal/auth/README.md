# Auth Module

This module provides authentication: registration with email verification, login and refresh token pairs (JWT access + opaque refresh tokens), and logout (revoke refresh tokens).

## Architecture and Responsibilities

The module follows the same layered architecture as the rest of the API (Handler, Service, Repository):

* **Registration:** Creates a user (via `user` package), links a credentials provider, and sends a verification email. The account cannot log in until the email is verified.
* **Email verification:** GET `/auth/verify-email?token=...` marks the user as verified using the token sent by email.
* **Resend verification:** Generates a new verification token and sends another email (e.g. via SES).
* **Login:** Validates email/password and returns an access token (JWT) plus a refresh token. Requires verified email.
* **Refresh:** Exchanges a valid refresh token for a new token pair (rotation). Invalid or expired refresh tokens are rejected.
* **Logout:** Revokes all refresh tokens for the authenticated user (uses `requestctx.UserOnly` like the user module).

## Data Model

### Tables

| Table | Description |
| :--- | :--- |
| `user_auth_providers` | Links users to providers (e.g. `credentials` with email as subject). UNIQUE(provider, provider_subject_id). |
| `refresh_tokens` | Stores SHA-256 hash of refresh tokens, user_id, expiry, revoked_at. Raw token is never stored. |

### Token behaviour

* **Access token:** JWT signed with HS256; contains user_id and email. Short-lived (e.g. 15 minutes).
* **Refresh token:** Opaque value, stored by hash. Long-lived (e.g. 7 days). Single use: after refresh, the old token is revoked.

## Error and HTTP Code Mapping

| Code | HTTP Status | When |
| :--- | :--- | :--- |
| `CodeInvalidRequestBody` | 400 | Invalid JSON or query binding. |
| `CodeValidationError` | 422 | Validation failed (with `details` per field where applicable). |
| `CodeInvalidCredentials` | 401 | Wrong email/password or invalid/expired refresh token. |
| `CodeTokenInvalid` | 401 | Invalid or expired refresh token (e.g. on refresh). |
| `CodeUnauthorized` | 401 | Logout without valid auth context. |
| `CodeEmailVerificationRequired` | 403 | Login or refresh with unverified email. |
| `CodeEmailAlreadyExists` | 409 | Register with existing email. |
| `CodeEmailAlreadyVerified` | 409 | Resend verification for already verified email. |
| `CodeInvalidVerificationToken` | 422 | Verify-email token missing, wrong or expired. |

## Technical Notes

* **Middleware:** Logout uses `requestctx.UserOnly(ctx)` to obtain the user ID from the request context (same pattern as user module).
* **Dependencies:** Service receives `UserRepository` (from user package), JWT secret, and optional `email.Sender` for verification emails.
* **Imports:** Standard → third-party → internal; `runtimeerror` is imported with alias `runtimeError` for readability.

## Tests

* **Handler tests (`handler_test.go`):** Success and error cases for Login, Refresh, Logout, Register, VerifyEmail, ResendVerification (validation, invalid credentials, email not verified, duplicate email, etc.). Use `SetupAuthHandlerTest(test *testing.T)` and `DecodeErrorResponse(test, body)`.
* **Service tests (`service_test.go`):** Login, Refresh, Register, VerifyEmail, token rotation and expiry.

Run: `go test ./internal/auth/...`
