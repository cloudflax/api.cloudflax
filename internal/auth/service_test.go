package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// En: hashTokenForTest calculates the SHA-256 hash of a test token.
// Es: hashTokenForTest calcula el hash SHA-256 de un token de prueba.
func hashTokenForTest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// En: setupServiceTest sets up the authentication service for tests.
// Es: setupServiceTest configura el servicio de autenticación para pruebas.
func setupServiceTest(test *testing.T) *Service {
	test.Helper()
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	return NewService(authRepository, userRepository, testJWTSecret, &noopEmailSender{})
}

// En: seedUser creates a test user.
// Es: seedUser crea un usuario para pruebas.
func seedUser(test *testing.T, name, email, password string) *user.User {
	test.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(test, u.SetPassword(password))
	require.NoError(test, database.DB.Create(u).Error)
	return u
}

// En: seedVerifiedUser creates a verified test user.
// Es: seedVerifiedUser crea un usuario con emailverifiedat establecido para que Login y Refresh tengan éxito.
func seedVerifiedUser(test *testing.T, name, email, password string) *user.User {
	test.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(test, u.SetPassword(password))
	require.NoError(test, database.DB.Create(u).Error)
	return u
}

// En: TestServiceLoginSuccess tests the successful login.
// Es: TestServiceLoginSuccess prueba el inicio de sesión exitoso.
func TestServiceLoginSuccess(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Alice", "alice@example.com", "password123")

	pair, err := service.Login("alice@example.com", "password123")
	require.NoError(test, err)
	assert.NotEmpty(test, pair.AccessToken)
	assert.NotEmpty(test, pair.RefreshToken)
	assert.False(test, pair.ExpiresAt.IsZero())
}

// En: TestServiceLoginInvalidCredentials tests the login with invalid credentials.
// Es: TestServiceLoginInvalidCredentials prueba el inicio de sesión con credenciales inválidas.
func TestServiceLoginInvalidCredentials(test *testing.T) {
	service := setupServiceTest(test)
	seedUser(test, "Bob", "bob@example.com", "correctpass")

	_, err := service.Login("bob@example.com", "wrongpass")
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}

// En: TestServiceLoginUnknownEmail tests the login with an unknown email.
// Es: TestServiceLoginUnknownEmail prueba el inicio de sesión con correo electrónico desconocido.
func TestServiceLoginUnknownEmail(test *testing.T) {
	service := setupServiceTest(test)

	_, err := service.Login("ghost@example.com", "password123")
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}

// En: TestServiceLoginEmailNotVerified tests the login with an unverified email.
// Es: TestServiceLoginEmailNotVerified prueba el inicio de sesión con correo electrónico no verificado.
func TestServiceLoginEmailNotVerified(test *testing.T) {
	service := setupServiceTest(test)
	seedUser(test, "Unverified", "unverified@example.com", "password123")

	_, err := service.Login("unverified@example.com", "password123")
	assert.ErrorIs(test, err, ErrEmailNotVerified)
}

// En: TestServiceValidateAccessTokenValid tests the validation of a valid access token.
// Es: TestServiceValidateAccessTokenValid prueba la validación de token de acceso válido.
func TestServiceValidateAccessTokenValid(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Carol", "carol@example.com", "password123")

	pair, err := service.Login("carol@example.com", "password123")
	require.NoError(test, err)

	userID, email, err := service.ValidateAccessToken(pair.AccessToken)
	require.NoError(test, err)
	assert.NotEmpty(test, userID)
	assert.Equal(test, "carol@example.com", email)
}

// En: TestServiceValidateAccessTokenInvalid tests the validation of an invalid access token.
// Es: TestServiceValidateAccessTokenInvalid prueba la validación de token de acceso inválido.
func TestServiceValidateAccessTokenInvalid(test *testing.T) {
	service := setupServiceTest(test)

	_, _, err := service.ValidateAccessToken("invalid.token.here")
	assert.Error(test, err)
}

// En: TestServiceValidateAccessTokenWrongSecret tests the validation of an access token with a wrong secret.
// Es: TestServiceValidateAccessTokenWrongSecret prueba la validación de token de acceso con secreto incorrecto.
func TestServiceValidateAccessTokenWrongSecret(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Dave", "dave@example.com", "password123")

	pair, err := service.Login("dave@example.com", "password123")
	require.NoError(test, err)

	otherService := NewService(NewRepository(database.DB), user.NewRepository(database.DB), "different-secret", &noopEmailSender{})
	_, _, err = otherService.ValidateAccessToken(pair.AccessToken)
	assert.Error(test, err)
}

// En: TestServiceRefreshTokensSuccess tests the successful refresh of tokens.
// Es: TestServiceRefreshTokensSuccess prueba el refresco de tokens exitoso.
func TestServiceRefreshTokensSuccess(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Eve", "eve@example.com", "password123")

	pair, err := service.Login("eve@example.com", "password123")
	require.NoError(test, err)

	newPair, err := service.RefreshTokens(pair.RefreshToken)
	require.NoError(test, err)
	assert.NotEmpty(test, newPair.AccessToken)
	assert.NotEmpty(test, newPair.RefreshToken)
	assert.NotEqual(test, pair.RefreshToken, newPair.RefreshToken)
}

// En: TestServiceRefreshTokensRotation tests the refresh of tokens with rotation.
// Es: TestServiceRefreshTokensRotation prueba el refresco de tokens con rotación.
func TestServiceRefreshTokensRotation(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Frank", "frank@example.com", "password123")

	pair, err := service.Login("frank@example.com", "password123")
	require.NoError(test, err)

	_, err = service.RefreshTokens(pair.RefreshToken)
	require.NoError(test, err)

	_, err = service.RefreshTokens(pair.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials, "reusing a rotated refresh token must fail")
}

// En: TestServiceRefreshTokensInvalidToken tests the refresh of tokens with an invalid token.
// Es: TestServiceRefreshTokensInvalidToken prueba el refresco de tokens con token inválido.
func TestServiceRefreshTokensInvalidToken(test *testing.T) {
	service := setupServiceTest(test)

	_, err := service.RefreshTokens("random-invalid-token")
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}

// En: TestServiceRefreshTokensEmailNotVerified tests the refresh of tokens with an unverified email.
// Es: TestServiceRefreshTokensEmailNotVerified prueba el refresco de tokens con correo electrónico no verificado.
func TestServiceRefreshTokensEmailNotVerified(test *testing.T) {
	service := setupServiceTest(test)
	u := seedUser(test, "Unverified", "unverified@example.com", "password123")

	// Manually create a refresh token for the unverified user (e.g. from before we required verification).
	rawToken := "test-refresh-token-for-unverified-user"
	hash := hashTokenForTest(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)
	require.NoError(test, service.repository.Create(&RefreshToken{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}))

	_, err := service.RefreshTokens(rawToken)
	assert.ErrorIs(test, err, ErrEmailNotVerified)
}

// En: TestServiceRegisterSuccess tests the successful registration.
// Es: TestServiceRegisterSuccess prueba el registro exitoso.
func TestServiceRegisterSuccess(test *testing.T) {
	service := setupServiceTest(test)

	u, token, err := service.Register("Alice", "alice@example.com", "password123")
	require.NoError(test, err)
	assert.NotEmpty(test, u.ID)
	assert.Equal(test, "alice@example.com", u.Email)
	assert.Nil(test, u.EmailVerifiedAt)
	assert.NotEmpty(test, token)

	var provider UserAuthProvider
	require.NoError(test, database.DB.Where("user_id = ? AND provider = ?", u.ID, ProviderCredentials).First(&provider).Error)
	assert.Equal(test, "alice@example.com", provider.ProviderSubjectID)
}

// En: TestServiceRegisterDuplicateEmail tests the registration with a duplicate email.
// Es: TestServiceRegisterDuplicateEmail prueba el registro con correo electrónico duplicado.
func TestServiceRegisterDuplicateEmail(test *testing.T) {
	service := setupServiceTest(test)

	_, _, err := service.Register("Alice", "dup@example.com", "password123")
	require.NoError(test, err)

	_, _, err = service.Register("Bob", "dup@example.com", "password456")
	assert.ErrorIs(test, err, user.ErrDuplicateEmail)
}

// En: TestServiceVerifyEmailSuccess tests the successful email verification.
// Es: TestServiceVerifyEmailSuccess prueba la verificación de correo electrónico exitosa.
func TestServiceVerifyEmailSuccess(test *testing.T) {
	service := setupServiceTest(test)

	_, token, err := service.Register("Carol", "carol@example.com", "password123")
	require.NoError(test, err)

	require.NoError(test, service.VerifyEmail(token))

	var u user.User
	require.NoError(test, database.DB.Where("email = ?", "carol@example.com").First(&u).Error)
	assert.NotNil(test, u.EmailVerifiedAt)
	assert.Nil(test, u.EmailVerificationToken)
}

// En: TestServiceVerifyEmailInvalidToken tests the email verification with an invalid token.
// Es: TestServiceVerifyEmailInvalidToken prueba la verificación de correo electrónico con token inválido.
func TestServiceVerifyEmailInvalidToken(test *testing.T) {
	service := setupServiceTest(test)

	err := service.VerifyEmail("00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(test, err, ErrInvalidVerificationToken)
}

// En: TestServiceResendVerificationSuccess tests the successful email verification resend.
// Es: TestServiceResendVerificationSuccess prueba el envío de correo de verificación exitoso.
func TestServiceResendVerificationSuccess(test *testing.T) {
	service := setupServiceTest(test)

	_, oldToken, err := service.Register("Dan", "dan@example.com", "password123")
	require.NoError(test, err)

	newToken, err := service.ResendVerification("dan@example.com")
	require.NoError(test, err)
	assert.NotEmpty(test, newToken)
	assert.NotEqual(test, oldToken, newToken)
}

// En: TestServiceResendVerificationAlreadyVerified tests the email verification resend with an already verified email.
// Es: TestServiceResendVerificationAlreadyVerified prueba el envío de correo de verificación con correo electrónico ya verificado.
func TestServiceResendVerificationAlreadyVerified(test *testing.T) {
	service := setupServiceTest(test)

	_, token, err := service.Register("Eve", "eve@example.com", "password123")
	require.NoError(test, err)
	require.NoError(test, service.VerifyEmail(token))

	_, err = service.ResendVerification("eve@example.com")
	assert.ErrorIs(test, err, ErrEmailAlreadyVerified)
}

// En: TestServiceLogoutRevokesAllTokens tests the logout and revocation of all tokens.
// Es: TestServiceLogoutRevokesAllTokens prueba el cierre de sesión y revocación de todos los tokens.
func TestServiceLogoutRevokesAllTokens(test *testing.T) {
	service := setupServiceTest(test)
	u := seedVerifiedUser(test, "Grace", "grace@example.com", "password123")

	pair1, err := service.Login("grace@example.com", "password123")
	require.NoError(test, err)
	pair2, err := service.Login("grace@example.com", "password123")
	require.NoError(test, err)

	require.NoError(test, service.Logout(u.ID))

	_, err = service.RefreshTokens(pair1.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials, "first session token should be revoked after logout")

	_, err = service.RefreshTokens(pair2.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials, "second session token should be revoked after logout")
}
